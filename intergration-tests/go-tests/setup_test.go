package go_test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/specularl2/specular/clients/geth/specular/bindings"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

const (
	//Layer1HttpUrl   = "http://localhost:8545"
	Layer2HttpUrl   = "http://localhost:4011"
	Layer1WssUrl    = "ws://localhost:8545"
	Layer2WssUrl    = "ws://localhost:4012"
	User1PrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	User2PrivateKey = "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	User3PrivateKey = "5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"

	DECIMAL1     = 1000000000000000000
	DECIMAL0_5   = 500000000000000000
	DECIMAL0_1   = 100000000000000000
	DECIMAL00_1  = 10000000000000000
	DECIMAL000_1 = 1000000000000000
)

var (
	l1Client *ethclient.Client
	l2Client *ethclient.Client
	l1ctx    context.Context
	l2ctx    context.Context

	User1Address = common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266")
	User2Address = common.HexToAddress("0x70997970c51812dc3a010c7d01b50e0d17dc79c8")
	User3Address = common.HexToAddress("0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc")

	Inbox         *bindings.ISequencerInbox
	Rollup        *bindings.IRollup
	InboxAddress  = common.HexToAddress("0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0")
	RollupAddress = common.HexToAddress("0x0165878A594ca255338adfa4d48449f69242Eb8F")
)

func init() {
	var err error
	l1ctx = context.Background()
	l2ctx = context.Background()
	if l1Client, err = ethclient.DialContext(l1ctx, Layer1WssUrl); err != nil {
		panic(err)
	}
	if l2Client, err = ethclient.DialContext(l2ctx, Layer2HttpUrl); err != nil {
		panic(err)
	}
	if Inbox, err = bindings.NewISequencerInbox(InboxAddress, l1Client); err != nil {
		panic(err)
	}
	if Rollup, err = bindings.NewIRollup(RollupAddress, l1Client); err != nil {
		panic(err)
	}
}

func mustGetL1Client(t *testing.T) *ethclient.Client {
	if l1Client == nil {
		t.Error("l1 client is nil")
	}
	return l1Client
}

func mustGetL2Client(t *testing.T) *ethclient.Client {
	if l2Client == nil {
		t.Error("l2 client is nil")
	}
	return l2Client
}

func buildAuth(t *testing.T, client *ethclient.Client, privateKey string, amount *big.Int) *bind.TransactOpts {
	privKey, err := crypto.HexToECDSA(privateKey)
	require.NoError(t, err)

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	require.NoError(t, err)

	//gasPrice, err := client.SuggestGasPrice(context.Background())
	//require.NoError(t, err)

	auth := bind.NewKeyedTransactor(privKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = amount             // in wei
	auth.GasLimit = uint64(3000000) // in units
	//auth.GasPrice = gasPrice
	auth.GasPrice = big.NewInt(1)
	return auth
}

func transferETH(t *testing.T, client *ethclient.Client, senderPrivKey string, fromAddress, toAddress common.Address, amount int64) {
	privateKey, err := crypto.HexToECDSA(senderPrivKey)
	require.NoError(t, err)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	require.NoError(t, err)

	value := big.NewInt(amount) // in wei (1 eth)
	gasLimit := uint64(21000)   // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	require.NoError(t, err)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	require.NoError(t, err)

	balance, err := client.BalanceAt(context.Background(), fromAddress, nil)
	t.Log(fmt.Sprintf("address: %s, balance: %s", fromAddress.String(), balance))
	err = client.SendTransaction(context.Background(), signedTx)
	require.NoError(t, err)

	t.Log("tx hash: ", signedTx.Hash())
}

func checkOrResetStakeAssertion(assertionID *big.Int, address common.Address, privateKey string) {

}

func createFakeAssertion(assertionID *big.Int, modifyOpIndex *big.Int) {

}
