package metrics

import (
	"context"
	"math"

	"github.com/flashbots/vpnham/config"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

const (
	metricsNamespace = "vpnham"
)

var (
	meter               otelapi.Meter
	latencyBoundariesUs otelapi.HistogramOption
)

func Setup(
	ctx context.Context,
	cfg *config.Metrics,
	observer func(context.Context, otelapi.Observer) error,
) error {
	for _, setup := range []func(context.Context, *config.Metrics) error{
		setupMeter,               // must come first
		setupLatencyBoundariesUs, // must come second
		setupBridgeActive,
		setupBridgeUp,
		setupTunnelInterfaceActive,
		setupTunnelInterfaceUp,
		setupProbesSent,
		setupProbesReturned,
		setupProbesFailed,
		setupProbesLatencyForward,
		setupProbesLatencyReturn,
	} {
		if err := setup(ctx, cfg); err != nil {
			return err
		}
	}

	// observables

	if _, err := meter.RegisterCallback(observer,
		BridgeActive,
		BridgeUp,
		TunnelInterfaceActive,
		TunnelInterfaceUp,
	); err != nil {
		return err
	}

	return nil
}

func setupMeter(ctx context.Context, _ *config.Metrics) error {
	res, err := resource.New(ctx)
	if err != nil {
		return err
	}

	exporter, err := prometheus.New(
		prometheus.WithNamespace(metricsNamespace),
		prometheus.WithoutScopeInfo(),
	)
	if err != nil {
		return err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)

	meter = provider.Meter(metricsNamespace)

	return nil
}

func setupLatencyBoundariesUs(ctx context.Context, cfg *config.Metrics) error {
	latencyBoundariesUs = otelapi.WithExplicitBucketBoundaries(func() []float64 {
		base := math.Exp(math.Log(float64(cfg.MaxLatencyUs)) / (float64(cfg.LatencyBucketsCount - 1)))
		res := make([]float64, 0, cfg.LatencyBucketsCount)
		for i := 0; i < cfg.LatencyBucketsCount; i++ {
			res = append(res,
				math.Round(2*math.Pow(base, float64(i)))/2,
			)
		}
		return res
	}()...)

	return nil
}

func setupBridgeActive(ctx context.Context, _ *config.Metrics) error {
	bridgeActive, err := meter.Int64ObservableGauge("bridge_active",
		otelapi.WithDescription("number of active bridges at a given moment"),
	)
	if err != nil {
		return err
	}
	BridgeActive = bridgeActive
	return nil
}

func setupBridgeUp(ctx context.Context, _ *config.Metrics) error {
	bridgeUp, err := meter.Int64ObservableGauge("bridge_up",
		otelapi.WithDescription("number of online bridges at a given moment"),
	)
	if err != nil {
		return err
	}
	BridgeUp = bridgeUp
	return nil
}

func setupTunnelInterfaceActive(ctx context.Context, _ *config.Metrics) error {
	tunnelInterfaceActive, err := meter.Int64ObservableGauge("tunnel_interface_active",
		otelapi.WithDescription("number of active tunnel interfaces at a given moment"),
	)
	if err != nil {
		return err
	}
	TunnelInterfaceActive = tunnelInterfaceActive
	return nil
}

func setupTunnelInterfaceUp(ctx context.Context, _ *config.Metrics) error {
	tunnelInterfaceUp, err := meter.Int64ObservableGauge("tunnel_interface_up",
		otelapi.WithDescription("number of online tunnel interfaces at a given moment"),
	)
	if err != nil {
		return err
	}
	TunnelInterfaceUp = tunnelInterfaceUp
	return nil
}

func setupProbesSent(ctx context.Context, _ *config.Metrics) error {
	probesSent, err := meter.Int64Counter("probes_sent",
		otelapi.WithDescription("counter for the sent probes"),
	)
	if err != nil {
		return err
	}
	ProbesSent = probesSent
	return nil
}

func setupProbesReturned(ctx context.Context, _ *config.Metrics) error {
	probesReturned, err := meter.Int64Counter("probes_returned",
		otelapi.WithDescription("counter for the returned probes"),
	)
	if err != nil {
		return err
	}
	ProbesReturned = probesReturned
	return nil
}

func setupProbesFailed(ctx context.Context, _ *config.Metrics) error {
	probesFailed, err := meter.Int64Counter("probes_failed",
		otelapi.WithDescription("counter for the failed probes"),
	)
	if err != nil {
		return err
	}
	ProbesFailed = probesFailed
	return nil
}

func setupProbesLatencyForward(ctx context.Context, _ *config.Metrics) error {
	probesLatencyForward, err := meter.Float64Histogram("probes_latency_forward",
		otelapi.WithDescription("latency of probes on the way there"),
		otelapi.WithUnit("us"),
		latencyBoundariesUs,
	)
	if err != nil {
		return err
	}
	ProbesLatencyForward = probesLatencyForward
	return nil
}

func setupProbesLatencyReturn(ctx context.Context, _ *config.Metrics) error {
	probesLatencyReturn, err := meter.Float64Histogram("probes_latency_return",
		otelapi.WithDescription("latency of probes on the way back"),
		otelapi.WithUnit("us"),
		latencyBoundariesUs,
	)
	if err != nil {
		return err
	}
	ProbesLatencyReturn = probesLatencyReturn
	return nil
}
