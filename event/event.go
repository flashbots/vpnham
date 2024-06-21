package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type Event interface {
	EventKind() string
	EventTimestamp() time.Time
}

type TunnelInterfaceEvent interface {
	Event
	TunnelInterface() string
}

type PartnerEvent interface {
	Event
	PartnerStatus() *types.BridgeStatus
}
