package config

import (
	"context"
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

func (m *Metrics) PostLoad(ctx context.Context) error {
	if m.ListenAddr == "" {
		m.ListenAddr = DefaultMetricsListenAddr
	}

	if m.LatencyBucketsCount == 0 {
		m.LatencyBucketsCount = DefaultLatencyBucketsCount
	}

	if m.MaxLatencyUs == 0 {
		m.MaxLatencyUs = DefaultMaxLatencyUs
	}

	return nil
}

func (m *Metrics) Validate(ctx context.Context) error {
	if err := m.ListenAddr.Validate(); err != nil {
		return fmt.Errorf("%w: %w",
			errMetricsInvalidListenAddr, err,
		)
	}
	return nil
}
