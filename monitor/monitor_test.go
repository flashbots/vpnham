package monitor_test

import (
	"testing"

	"github.com/flashbots/vpnham/monitor"
	"github.com/stretchr/testify/assert"
)

type event struct {
	sequence uint64
	probe    monitor.Status
	expected monitor.Status
}

type testCase = []event

func TestMonitor(t *testing.T) {
	m, err := monitor.New(5, 2)
	assert.NoError(t, err)

	tc := testCase{
		{sequence: 0, probe: monitor.Pending, expected: monitor.Pending},
		{sequence: 1, probe: monitor.Pending, expected: monitor.Pending},
		{sequence: 2, probe: monitor.Up, expected: monitor.Pending},
		{sequence: 3, probe: monitor.Down, expected: monitor.Pending},
		{sequence: 4, probe: monitor.Down, expected: monitor.Pending},
		{sequence: 5, probe: monitor.Down, expected: monitor.Pending},
		{sequence: 6, probe: monitor.Down, expected: monitor.Pending},
		{sequence: 7, probe: monitor.Down, expected: monitor.Down},
		{sequence: 8, probe: monitor.Pending, expected: monitor.Down},
		{sequence: 9, probe: monitor.Pending, expected: monitor.Down},
		{sequence: 9, probe: monitor.Up, expected: monitor.Pending},
		{sequence: 8, probe: monitor.Up, expected: monitor.Up},
		{sequence: 10, probe: monitor.Pending, expected: monitor.Up},
	}

	for idx, step := range tc {
		m.RegisterStatus(step.sequence, step.probe)
		assert.Equal(t, step.expected, m.Status(), "step %d: expected %s, got %s",
			idx, step.expected, m.Status(),
		)
	}
}
