package go_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/specularl2/specular/clients/geth/specular/bindings"
	"github.com/stretchr/testify/require"
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
	//// stake the given assertion
	//assertionID := big.NewInt(0)
	//checkOrResetStakeAssertion(assertionID, User3Address, User3PrivateKey)
	//
	//// get latest assertionID and create assertion
	//states, vmHash := createFakeAssertion()
	//
	//var timeoutTK = time.NewTicker(time.Second * 10)
	//
	//// subscribe assertion
	//// subscribe CreateAssertion called
	//createdCh := make(chan *bindings.IRollupAssertionCreated, 64)
	//createdSub, err := Rollup.WatchAssertionCreated(&bind.WatchOpts{}, createdCh)
	//require.NoError(t, err)
	//defer createdSub.Unsubscribe()
	//
	//// submit new assertion
	//Rollup.CreateAssertion()
	//
	//select {
	//case assertionCreateEvent := <-createdCh:
	//	t.Log(assertionCreateEvent)
	//case <-timeoutTK.C:
	//	t.Error("step AssertionCreated timeout")
	//}
	//
	//// call challenge assertion
	//Rollup.ChallengeAssertion()
	//
	//// subscribe assertion challenged
	//challengedCh := make(chan *bindings.IRollupAssertionChallenged, 64)
	//challengedSub, err := Rollup.WatchAssertionChallenged(&bind.WatchOpts{}, challengedCh)
	//require.NoError(t, err)
	//defer challengedSub.Unsubscribe()
	//select {
	//case assertionChallengedEvent := <-challengedCh:
	//	t.Log(assertionChallengedEvent)
	//case <-timeoutTK.C:
	//	t.Error("step AssertionChallenged timeout")
	//}
	//
	//// subscribe bisected
	//
	//// subscribe challenge completed
	//
	//// respond bisection and completed challenge
}
