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

	DefaultAWSTimeout     = 15 * time.Second
	DefaultGCPTimeout     = 15 * time.Second
	DefaultScriptsTimeout = 30 * time.Second

	DefaultRouteIDPrefix = "vpnham"

	DefaultGCPRoutePriority uint32 = 1000

	DefaultMetricsListenAddr = "0.0.0.0:8000"

	DefaultLatencyBucketsCount = 33      // from 1us to 1s
	DefaultMaxLatencyUs        = 1000000 // 1s

	DefaultReapplyFactor = 2.0
)
