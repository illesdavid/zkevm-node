package e2e

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPolygonHermez/zkevm-node/hex"
	"github.com/0xPolygonHermez/zkevm-node/test/contracts/bin/Counter"
	"github.com/0xPolygonHermez/zkevm-node/test/contracts/bin/Creates"
	"github.com/0xPolygonHermez/zkevm-node/test/contracts/bin/ERC20"
	"github.com/0xPolygonHermez/zkevm-node/test/contracts/bin/EmitLog"
	"github.com/0xPolygonHermez/zkevm-node/test/contracts/bin/Revert2"
	"github.com/0xPolygonHermez/zkevm-node/test/operations"
	"github.com/0xPolygonHermez/zkevm-node/test/testutils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

const fixedTxGasLimit uint64 = 100000

func createEthTransferSignedTx(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client, customData map[string]interface{}) (*ethTypes.Transaction, error) {
	nonce, err := client.PendingNonceAt(ctx, auth.From)
	require.NoError(t, err)

	gasPrice, err := client.SuggestGasPrice(ctx)
	require.NoError(t, err)

	to := common.HexToAddress("0x1275fbb540c8efc58b812ba83b0d0b8b9917ae98")

	gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: auth.From,
		To:   &to,
	})
	require.NoError(t, err)

	tx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		GasPrice: gasPrice,
		Gas:      gas,
	})

	return auth.Signer(auth.From, tx)
}

func createScDeploySignedTx(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client, customData map[string]interface{}) (*ethTypes.Transaction, error) {
	nonce, err := client.PendingNonceAt(ctx, auth.From)
	require.NoError(t, err)

	gasPrice, err := client.SuggestGasPrice(ctx)
	require.NoError(t, err)

	scByteCode, err := testutils.ReadBytecode("Counter/Counter.bin")
	require.NoError(t, err)
	data := common.Hex2Bytes(scByteCode)

	gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: auth.From,
		Data: data,
	})
	require.NoError(t, err)

	tx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gas,
		Data:     data,
	})

	return auth.Signer(auth.From, tx)
}

func prepareScCall(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client) (map[string]interface{}, error) {
	_, tx, sc, err := EmitLog.DeployEmitLog(auth, client)
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, tx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	return map[string]interface{}{
		"sc": sc,
	}, nil
}

func createScCallSignedTx(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client, customData map[string]interface{}) (*ethTypes.Transaction, error) {
	scInterface := customData["sc"]
	sc := scInterface.(*EmitLog.EmitLog)

	opts := *auth
	opts.NoSend = true

	tx, err := sc.EmitLogs(&opts)
	require.NoError(t, err)

	return tx, nil
}

func prepareERC20Transfer(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client) (map[string]interface{}, error) {
	_, tx, sc, err := ERC20.DeployERC20(auth, client, "MyToken", "MT")
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, tx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	tx, err = sc.Mint(auth, big.NewInt(1000000000))
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, tx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	return map[string]interface{}{
		"sc": sc,
	}, nil
}

func createERC20TransferSignedTx(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client, customData map[string]interface{}) (*ethTypes.Transaction, error) {
	scInterface := customData["sc"]
	sc := scInterface.(*ERC20.ERC20)

	opts := *auth
	opts.NoSend = true

	to := common.HexToAddress("0x1275fbb540c8efc58b812ba83b0d0b8b9917ae98")

	tx, err := sc.Transfer(&opts, to, big.NewInt(123456))
	require.NoError(t, err)

	return tx, nil
}

func createScDeployRevertedSignedTx(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client, customData map[string]interface{}) (*ethTypes.Transaction, error) {
	nonce, err := client.PendingNonceAt(ctx, auth.From)
	require.NoError(t, err)

	gasPrice, err := client.SuggestGasPrice(ctx)
	require.NoError(t, err)

	scByteCode, err := testutils.ReadBytecode("Revert/Revert.bin")
	require.NoError(t, err)
	data := common.Hex2Bytes(scByteCode)

	tx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      fixedTxGasLimit,
		Data:     data,
	})

	return auth.Signer(auth.From, tx)
}

func prepareScCallReverted(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client) (map[string]interface{}, error) {
	_, tx, sc, err := Revert2.DeployRevert2(auth, client)
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, tx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	return map[string]interface{}{
		"sc": sc,
	}, nil
}

func createScCallRevertedSignedTx(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client, customData map[string]interface{}) (*ethTypes.Transaction, error) {
	scInterface := customData["sc"]
	sc := scInterface.(*Revert2.Revert2)

	opts := *auth
	opts.NoSend = true
	opts.GasLimit = fixedTxGasLimit

	tx, err := sc.GenerateError(&opts)
	require.NoError(t, err)

	return tx, nil
}

func prepareERC20TransferReverted(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client) (map[string]interface{}, error) {
	_, tx, sc, err := ERC20.DeployERC20(auth, client, "MyToken2", "MT2")
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, tx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, tx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	return map[string]interface{}{
		"sc": sc,
	}, nil
}

func createERC20TransferRevertedSignedTx(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client, customData map[string]interface{}) (*ethTypes.Transaction, error) {
	scInterface := customData["sc"]
	sc := scInterface.(*ERC20.ERC20)

	opts := *auth
	opts.NoSend = true
	opts.GasLimit = fixedTxGasLimit

	to := common.HexToAddress("0x1275fbb540c8efc58b812ba83b0d0b8b9917ae98")

	tx, err := sc.Transfer(&opts, to, big.NewInt(123456))
	require.NoError(t, err)

	return tx, nil
}

func prepareCreate(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client) (map[string]interface{}, error) {
	_, tx, sc, err := Creates.DeployCreates(auth, client)
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, tx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	return map[string]interface{}{
		"sc": sc,
	}, nil
}

func createCreateSignedTx(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client, customData map[string]interface{}) (*ethTypes.Transaction, error) {
	scInterface := customData["sc"]
	sc := scInterface.(*Creates.Creates)

	opts := *auth
	opts.NoSend = true

	byteCode := hex.DecodeBig(Counter.CounterBin).Bytes()

	tx, err := sc.OpCreate(&opts, byteCode, big.NewInt(0).SetInt64(int64(len(byteCode))))
	require.NoError(t, err)

	return tx, nil
}

func createCreate2SignedTx(t *testing.T, ctx context.Context, auth *bind.TransactOpts, client *ethclient.Client, customData map[string]interface{}) (*ethTypes.Transaction, error) {
	scInterface := customData["sc"]
	sc := scInterface.(*Creates.Creates)

	opts := *auth
	opts.NoSend = true

	byteCode := hex.DecodeBig(Counter.CounterBin).Bytes()

	tx, err := sc.OpCreate2(&opts, byteCode, big.NewInt(0).SetInt64(int64(len(byteCode))))
	require.NoError(t, err)

	return tx, nil
}
