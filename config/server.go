package config

import (
	"context"
	"fmt"
)

type Server struct {
	Bridges map[string]*Bridge `yaml:"bridges"`
	Metrics *Metrics           `yaml:"metrics"`
}

func (s *Server) PostLoad(ctx context.Context) error {
	// bridges

	for bn, b := range s.Bridges {
		b.Name = bn

		if err := b.PostLoad(ctx); err != nil {
			return err
		}
	}

	// metrics

	if err := s.Metrics.PostLoad(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Server) Validate(ctx context.Context) error {
	for name, bridge := range s.Bridges {
		if err := bridge.Validate(ctx); err != nil {
			return fmt.Errorf("%s: %w",
				name, err,
			)
		}
	}

	if err := s.Metrics.Validate(ctx); err != nil {
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
