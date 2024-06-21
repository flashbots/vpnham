package config

import "time"

const (
	DefaultConfigFile = ".vpnham.yaml"

	DefaultLogLevel = "info"
	DefaultLogMode  = "prod"

	DefaultPartnerStatusTimeout = time.Second
	DefaultProbeInterval        = 15 * time.Second

	DefaultThresholdDown = 5
	DefaultThresholdUp   = 2

	DefaultScriptsTimeout = 30 * time.Second

	DefaultMetricsListenAddr   = "0.0.0.0:8000"
	DefaultLatencyBucketsCount = 33      // from 1us to 1s
	DefaultMaxLatencyUs        = 1000000 // 1s
)
