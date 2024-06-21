package config

import (
	"errors"
	"fmt"

	"github.com/flashbots/vpnham/types"
)

type Metrics struct {
	ListenAddr types.Address `yaml:"listen_addr"`

	LatencyBucketsCount int `yaml:"latency_buckets_count"`
	MaxLatencyUs        int `yaml:"max_latency_us"`
}

var (
	errMetricsInvalidListenAddr = errors.New("invalid metrics listen addr")
)

func (m *Metrics) Validate() error {
	if err := m.ListenAddr.Validate(); err != nil {
		return fmt.Errorf("%w: %w",
			errMetricsInvalidListenAddr, err,
		)
	}
	return nil
}
