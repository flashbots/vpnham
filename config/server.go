package config

import (
	"fmt"
)

type Server struct {
	Bridges        map[string]*Bridge `yaml:"bridges"`
	DefaultScripts *Scripts           `yaml:"default_scripts"`
	Metrics        *Metrics           `yaml:"metrics"`
}

func (s *Server) PostLoad() {
	{ // bridges
		for bn, b := range s.Bridges {
			b.Name = bn

			if b.PartnerStatusTimeout == 0 {
				b.PartnerStatusTimeout = DefaultPartnerStatusTimeout
			}

			if b.ProbeInterval == 0 {
				b.ProbeInterval = DefaultProbeInterval
			}

			if b.PartnerStatusThresholdDown == 0 {
				b.PartnerStatusThresholdDown = DefaultThresholdDown
			}

			if b.PartnerStatusThresholdUp == 0 {
				b.PartnerStatusThresholdUp = DefaultThresholdUp
			}

			{ // interfaces
				for ifsName, ifs := range b.TunnelInterfaces {
					ifs.Name = ifsName

					if ifs.ThresholdDown == 0 {
						ifs.ThresholdDown = DefaultThresholdDown
					}

					if ifs.ThresholdUp == 0 {
						ifs.ThresholdUp = DefaultThresholdUp
					}
				}
			}

			{ // scripts
				if b.Scripts == nil {
					b.Scripts = &Scripts{}
				}
				if b.Scripts.BridgeActivate == nil {
					if s.DefaultScripts != nil {
						b.Scripts.BridgeActivate = s.DefaultScripts.BridgeActivate
					}
				}
				if b.Scripts.InterfaceActivate == nil {
					if s.DefaultScripts != nil {
						b.Scripts.InterfaceActivate = s.DefaultScripts.InterfaceActivate
					}
				}
				if b.Scripts.InterfaceDeactivate == nil {
					if s.DefaultScripts != nil {
						b.Scripts.InterfaceDeactivate = s.DefaultScripts.InterfaceDeactivate
					}
				}
				if b.ScriptsTimeout == 0 {
					b.ScriptsTimeout = DefaultScriptsTimeout
				}
			}
		}
	}

	{ // metrics
		if s.Metrics.ListenAddr == "" {
			s.Metrics.ListenAddr = DefaultMetricsListenAddr
		}

		if s.Metrics.LatencyBucketsCount == 0 {
			s.Metrics.LatencyBucketsCount = DefaultLatencyBucketsCount
		}

		if s.Metrics.MaxLatencyUs == 0 {
			s.Metrics.MaxLatencyUs = DefaultMaxLatencyUs
		}
	}
}

func (s *Server) Validate() error {
	for name, bridge := range s.Bridges {
		if err := bridge.Validate(); err != nil {
			return fmt.Errorf("%s: %w",
				name, err,
			)
		}
	}

	if err := s.Metrics.Validate(); err != nil {
		return err
	}

	return nil
}

func (s *Server) BridgesCount() int {
	return len(s.Bridges)
}

func (s *Server) TunnelInterfacesCount() int {
	c := 0
	for _, b := range s.Bridges {
		c += b.TunnelInterfacesCount()
	}
	return c
}

func (s *Server) EventSourcesCount() int {
	return 1 + s.BridgesCount() + s.TunnelInterfacesCount()
}
