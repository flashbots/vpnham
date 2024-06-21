package monitor

import (
	"errors"
	"fmt"
)

const (
	maxThreshold = 10
)

type Monitor struct {
	dnThreshold, upThreshold int

	sequence uint64

	history []Status
}

var (
	errDownThresholdIsInvalid = errors.New("down threshold is invalid")
	errUpThresholdIsInvalid   = errors.New("up threshold is invalid")
	errSequenceTooHigh        = errors.New("sequence is too high")
)

func New(downThreshold, upThreshold int) (*Monitor, error) {
	if downThreshold < 2 || maxThreshold < downThreshold {
		return nil, fmt.Errorf("%w: expected 1 < N < %d, got %d",
			errDownThresholdIsInvalid, maxThreshold, downThreshold,
		)
	}

	if upThreshold < 2 || maxThreshold < upThreshold {
		return nil, fmt.Errorf("%w: expected 1 < N < %d, got %d",
			errUpThresholdIsInvalid, maxThreshold, upThreshold,
		)
	}

	history := make([]Status, max(downThreshold, upThreshold)+1)
	for idx := len(history); idx < cap(history); idx++ {
		history = append(history, 0)
	}

	return &Monitor{
		dnThreshold: downThreshold,
		upThreshold: upThreshold,

		history: history,
	}, nil
}

func (m *Monitor) Sequence() uint64 {
	return m.sequence
}

func (m *Monitor) advanceSequence(sequence uint64) {
	jump := int(sequence - m.sequence)

	if jump >= len(m.history) {
		for idx := 0; idx < len(m.history); idx++ {
			m.history[idx] = Pending
		}
		return
	}

	copy(m.history, m.history[jump:])
	for idx := len(m.history) - jump; idx < len(m.history); idx++ {
		m.history[idx] = Pending
	}

	m.sequence = sequence
}

func (m *Monitor) RegisterStatus(sequence uint64, status Status) {
	if sequence > m.sequence {
		m.advanceSequence(sequence)
	}

	offset := m.sequence - sequence
	historyLength := uint64(len(m.history))

	if offset > historyLength {
		// already past the history window
		return
	}

	historyStart := m.sequence - historyLength + 1
	index := sequence - historyStart

	m.history[int(index)] = status
}

func (m *Monitor) Status() Status {
	dnStreak := 0
	unStreak := 0
	upStreak := 0

	window := min(len(m.history), max(m.dnThreshold, m.upThreshold))

	firstState := true
	for idx := len(m.history) - 1; idx >= len(m.history)-window; idx-- {
		switch m.history[idx] {
		case Up:
			dnStreak = -1
			unStreak = -1
			if upStreak >= 0 {
				upStreak++
			}
		case Down:
			if dnStreak >= 0 {
				dnStreak++
			}
			unStreak = -1
			upStreak = -1
		case Pending:
			if dnStreak >= 0 {
				dnStreak++ // sequence of unknown states also counts as down-streak
			}
			if unStreak >= 0 {
				unStreak++
			}
			if firstState && upStreak >= 0 {
				upStreak++ // first unknown state also counts as up-streak
			}
		}
		firstState = false

		if dnStreak >= m.dnThreshold && unStreak < dnStreak {
			return Down
		}
		if upStreak >= m.upThreshold && unStreak < upStreak {
			return Up
		}
	}

	return Pending
}
