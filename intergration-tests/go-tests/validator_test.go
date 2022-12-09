package go_test

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/specularl2/specular/clients/geth/specular/bindings"
	"github.com/stretchr/testify/require"
	"math/big"
	"sync"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	// build layer2 tx and submit
	wg := sync.WaitGroup{}
	txNum := 1
	wg.Add(txNum)
	go func() {
		txCount := 0
		for {
			if txCount >= txNum {
				return
			}
			txCount++
			transferETH(t, mustGetL2Client(t), User1PrivateKey, User1Address, User2Address, DECIMAL000_1)
			wg.Done()
		}
	}()

	var timeout5TK = time.NewTicker(time.Second * 5)
	var timeout10TK = time.NewTicker(time.Second * 10)
	//var timeout20TK = time.NewTicker(time.Second * 20)
	//var timeout30TK = time.NewTicker(time.Second * 20)

	// subscribe AppendTxBatch called
	appendedCh := make(chan *bindings.ISequencerInboxTxBatchAppended, 64)
	appendedSub, err := Inbox.WatchTxBatchAppended(&bind.WatchOpts{Context: l1ctx}, appendedCh)
	require.NoError(t, err)
	defer appendedSub.Unsubscribe()
	select {
	case txBatchAppendedEvent := <-appendedCh:
		t.Log(fmt.Sprintf("\ntxBatchAppendedEvent\n batchNum: %s\n startTxNum: %s\n endTxNum: %s",
			txBatchAppendedEvent.BatchNumber, txBatchAppendedEvent.StartTxNumber, txBatchAppendedEvent.EndTxNumber))
	case <-timeout5TK.C:
		t.Error("step TxBatchAppended timeout")
	}

	// subscribe AdvanceStake called
	advancedCh := make(chan *bindings.IRollupAdvanceStake, 64)
	advancedSub, err := Rollup.WatchAdvanceStake(&bind.WatchOpts{Context: l1ctx}, advancedCh)
	require.NoError(t, err)
	defer advancedSub.Unsubscribe()
	select {
	case advanceStakeEvent := <-advancedCh:
		t.Log(fmt.Sprintf("\nadvanceStakeEvent\n AssertionID: %s", advanceStakeEvent.AssertionID.String()))
	case <-timeout10TK.C:
		t.Error("step AdvanceStake timeout")
	}
}

func TestChallenge(t *testing.T) {
	// auto stake to the latest assertion
	checkOrResetStakeAssertion(t, User3Address, User3PrivateKey)

	var timeout5TK = time.NewTicker(time.Second * 5)
	var timeout10TK = time.NewTicker(time.Second * 10)
	//var timeout20TK = time.NewTicker(time.Second * 20)
	//var timeout30TK = time.NewTicker(time.Second * 20)

	// subscribe assertion
	// subscribe CreateAssertion called
	createdCh := make(chan *bindings.IRollupAssertionCreated, 64)
	createdSub, err := Rollup.WatchAssertionCreated(&bind.WatchOpts{Context: l1ctx}, createdCh)
	require.NoError(t, err)
	defer createdSub.Unsubscribe()

	// subscribe assertion challenged
	challengedCh := make(chan *bindings.IRollupAssertionChallenged, 64)
	challengedSub, err := Rollup.WatchAssertionChallenged(&bind.WatchOpts{Context: l1ctx}, challengedCh)
	require.NoError(t, err)
	defer challengedSub.Unsubscribe()

	bisectedCh := make(chan *bindings.IChallengeBisected, 64)
	bisectedSub, err := Challenge.WatchBisected(&bind.WatchOpts{Context: l1ctx}, bisectedCh)
	require.NoError(t, err)
	defer bisectedSub.Unsubscribe()

	challengeCompletedCh := make(chan *bindings.IChallengeChallengeCompleted, 64)
	challengeCompletedSub, err := Challenge.WatchChallengeCompleted(&bind.WatchOpts{Context: l1ctx}, challengeCompletedCh)
	require.NoError(t, err)
	defer challengeCompletedSub.Unsubscribe()

	// submit new tx
	transferETH(t, mustGetL2Client(t), User1PrivateKey, User1Address, User2Address, DECIMAL000_1)

	// wait for sequencer propose a new assertion and submit a challenge
	var fakeAssertion *bindings.IRollupAssertionCreated
	var assertionCreateEvent *bindings.IRollupAssertionCreated
	//var states []*proof.ExecutionState
	select {
	case assertionCreateEvent = <-createdCh:
		t.Log(fmt.Sprintf("\nassertionCreateEvent\n AssertionID: %s\n AsserterAddr: %s\n VmHash: %s\n InboxSize: %s\n L2GasUsed: %s\n",
			assertionCreateEvent.AssertionID, assertionCreateEvent.AsserterAddr.String(), hex.EncodeToString(assertionCreateEvent.VmHash[:]),
			assertionCreateEvent.InboxSize.String(), assertionCreateEvent.L2GasUsed.String()))

		// get the latest assertion and create a fake assertion
		fakeAssertion = createFakeAssertion(assertionCreateEvent)

		// submit new assertion
		tx, err := Rollup.CreateAssertion(&bind.TransactOpts{Context: l1ctx},
			fakeAssertion.VmHash, fakeAssertion.InboxSize, fakeAssertion.L2GasUsed,
			assertionCreateEvent.VmHash, assertionCreateEvent.L2GasUsed)
		require.NoError(t, err)
		t.Log("tx hash: ", tx.Hash())

		// call challenge assertion
		tx, err = Rollup.ChallengeAssertion(
			&bind.TransactOpts{
				Context: l1ctx,
			},
			[2]common.Address{
				User1Address,
				User3Address,
			},
			[2]*big.Int{
				assertionCreateEvent.AssertionID,
				fakeAssertion.AssertionID,
			},
		)
		require.NoError(t, err)
		t.Log("tx hash: ", tx.Hash())

	case <-timeout5TK.C:
		t.Error("step AssertionCreated timeout")
	}

	select {
	case assertionChallengedEvent := <-challengedCh:
		t.Log(fmt.Sprintf("\nassertionChallengedEvent\n AssertionID: %s\n ChallengeAddr: %s\n",
			assertionChallengedEvent.AssertionID.String(), assertionChallengedEvent.ChallengeAddr.String()))

		// get the latest assertion state
		//recentAssertion, err := AssertionMap.Assertions(&bind.CallOpts{Context: l1ctx},
		//	new(big.Int).Sub(assertionChallengedEvent.AssertionID, big.NewInt(1)))
		//require.NoError(t, err)

		//// create offline state
		//states, err := proof.GenerateStates(
		//	v.ProofBackend, // TODO
		//	l1ctx,
		//	recentAssertion.L2GasUsed, // TODO
		//	fakeAssertion.InboxSize.Sub(fakeAssertion.InboxSize, big.NewInt(1)).Uint64(),
		//	fakeAssertion.InboxSize.Uint64(),
		//	nil,
		//)
		require.NoError(t, err)
	case <-timeout10TK.C:
		t.Error("step AssertionChallenged timeout")
	}

	for {
		var timeoutTK = time.NewTicker(time.Second * 10)
		select {
		case bisectedEvent := <-bisectedCh:
			t.Log(bisectedEvent)
			// get responder
			responder, err := Challenge.CurrentResponder(&bind.CallOpts{Context: l1ctx})
			require.NoError(t, err)
			t.Log("responder is: ", responder.String())
			if responder == User3Address {
				//abi, err := bindings.IChallengeMetaData.GetAbi()
				//require.NoError(t, err)
				//transactOpts := buildAuth(t, l1Client, User3PrivateKey, big.NewInt(0))
				//challengeSession := &bindings.IChallengeSession{
				//	Contract:     Challenge,
				//	CallOpts:     bind.CallOpts{Pending: true, Context: l1ctx},
				//	TransactOpts: *transactOpts,
				//}
				//// TODO
				//base, err := services.NewBaseService(eth, proofBackend, cfg, auth)
				//require.NoError(t, err)
				//err = services.RespondBisection(base, abi, challengeSession, bisectedEvent, states, assertionCreateEvent.VmHash, false)
				t.Log(err)
			}
		case <-timeoutTK.C:
			t.Error("step AssertionChallenged timeout")
		}
	}

	// respond bisection and completed challenge
}
