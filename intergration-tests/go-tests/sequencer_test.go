package go_test

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/specularl2/specular/clients/geth/specular/bindings"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestRollup(t *testing.T) {
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
	wg.Wait()

	var timeout5TK = time.NewTicker(time.Second * 5)
	var timeout10TK = time.NewTicker(time.Second * 10)
	//var timeout20TK = time.NewTicker(time.Second * 20)
	var timeout30TK = time.NewTicker(time.Second * 20)

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

	// subscribe CreateAssertion called
	createdCh := make(chan *bindings.IRollupAssertionCreated, 64)
	createdSub, err := Rollup.WatchAssertionCreated(&bind.WatchOpts{Context: l1ctx}, createdCh)
	require.NoError(t, err)
	defer createdSub.Unsubscribe()
	select {
	case assertionCreateEvent := <-createdCh:
		t.Log(fmt.Sprintf("\nassertionCreateEvent\n AssertionID: %s\n AsserterAddr: %s\n VmHash: %s\n InboxSize: %s\n L2GasUsed: %s\n",
			assertionCreateEvent.AssertionID, assertionCreateEvent.AsserterAddr.String(), hex.EncodeToString(assertionCreateEvent.VmHash[:]),
			assertionCreateEvent.InboxSize.String(), assertionCreateEvent.L2GasUsed.String()))
	case <-timeout10TK.C:
		t.Error("step AssertionCreated timeout")
	}

	// subscribe ConfirmFirstUnresolvedAssertion called
	confirmedCh := make(chan *bindings.IRollupAssertionConfirmed, 64)
	confirmedSub, err := Rollup.WatchAssertionConfirmed(&bind.WatchOpts{Context: l1ctx}, confirmedCh)
	require.NoError(t, err)
	defer confirmedSub.Unsubscribe()
	select {
	case assertionConfirmedEvent := <-confirmedCh:
		t.Log(fmt.Sprintf("\nassertionConfirmedEvent\n AssertionID: %s", assertionConfirmedEvent.AssertionID))
	case <-timeout30TK.C:
		t.Error("step AssertionConfirmed timeout")
	}
}

func TestRespondChallenge(t *testing.T) {
	//// build layer2 tx and submit
	//go func() {
	//	txCount := 0
	//	for {
	//		txCount++
	//		transferETH(t, mustGetL2Client(t), User1PrivateKey, User1Address, User2Address, DECIMAL000_1)
	//		time.Sleep(1)
	//		if txCount > 10 {
	//			return
	//		}
	//	}
	//}()
	//
	//var timeoutTK = time.NewTicker(time.Second * 10)
	//
	//// subscribe AppendTxBatch called
	//appendedCh := make(chan *bindings.ISequencerInboxTxBatchAppended, 64)
	//appendedSub, err := Inbox.WatchTxBatchAppended(&bind.WatchOpts{}, appendedCh)
	//require.NoError(t, err)
	//defer appendedSub.Unsubscribe()
	//select {
	//case txBatchAppendedEvent := <-appendedCh:
	//	t.Log(txBatchAppendedEvent)
	//case <-timeoutTK.C:
	//	t.Error("step TxBatchAppended timeout")
	//}
	//
	//// create assertion
	//vmHash := ""
	//
	////
}
