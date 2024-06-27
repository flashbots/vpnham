package event

import "github.com/flashbots/vpnham/types"

type PartnerPollEvent interface {
	Event
	PartnerStatus() *types.BridgeStatus
}
