package coregrpc_test

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/mempool"

	"github.com/tendermint/tendermint/abci/example/kvstore"
	core_grpc "github.com/tendermint/tendermint/rpc/grpc"
	rpctest "github.com/tendermint/tendermint/rpc/test"
)

func TestMain(m *testing.M) {
	// start a tendermint node in the background to test against
	app := kvstore.NewApplication()
	node := rpctest.StartTendermint(app)

	code := m.Run()

	// and shut down proper at the end
	rpctest.StopTendermint(node)
	os.Exit(code)
}

func TestBroadcastTx(t *testing.T) {
	txdata, err := json.Marshal(mempool.ExTxInfo{GasPrice: big.NewInt(1)})
	require.NoError(t, err)
	res, err := rpctest.GetGRPCClient().BroadcastTx(
		context.Background(),
		&core_grpc.RequestBroadcastTx{Tx: txdata},
	)
	require.NoError(t, err)
	require.EqualValues(t, 0, res.CheckTx.Code)
	require.EqualValues(t, 0, res.DeliverTx.Code)
}
