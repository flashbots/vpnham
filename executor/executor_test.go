package executor_test

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/flashbots/vpnham/event"
// 	"github.com/flashbots/vpnham/executor"
// 	"github.com/flashbots/vpnham/logutils"
// 	"github.com/flashbots/vpnham/types"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/zap"
// )

// func TestExecutor(t *testing.T) {
// 	l, err := zap.NewDevelopment()
// 	assert.NoError(t, err)

// 	ctx := logutils.ContextWithLogger(context.Background(), l)

// 	zap.ReplaceGlobals(l)

// 	bridgeActivate := types.Script{}
// 	interfaceActivate := types.Script{}
// 	interfaceDeactivate := types.Script{}

// 	ex, err := executor.New(
// 		"test",
// 		uuid.Must(uuid.NewRandom()),
// 		bridgeActivate,
// 		interfaceActivate,
// 		interfaceDeactivate,
// 	)
// 	assert.NoError(t, err)

// 	ex.Run(ctx, nil)

// 	ex.ExecuteBridgeActivate(ctx, &event.BridgeActivated{
// 		BridgeInterface: "lo0",
// 		BridgePeerCIDR:  "10.0.0.0/24",
// 		Timestamp:       time.Now(),
// 	})

// 	ex.ExecuteInterfaceActivate(ctx, &event.TunnelInterfaceActivated{
// 		BridgeInterface: "lo0",
// 		BridgePeerCIDR:  "10.0.0.0/24",
// 		Interface:       "lo0",
// 		Timestamp:       time.Time{},
// 	})

// 	ex.ExecuteInterfaceDeactivate(ctx, &event.TunnelInterfaceDeactivated{
// 		BridgeInterface: "lo0",
// 		BridgePeerCIDR:  "10.0.0.0/24",
// 		Interface:       "lo0",
// 		Timestamp:       time.Time{},
// 	})

// 	time.Sleep(time.Second)
// }
