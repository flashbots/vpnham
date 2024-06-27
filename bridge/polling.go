package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/types"
	"go.uber.org/zap"
)

var (
	errPartnerBridgeNameIsDifferent = errors.New("partner's bridge name is different from ours")
	errPartnerRoleIsIdentical       = errors.New("partner has the role identical to ours")
)

func (s *Server) pollPartnerBridge(ctx context.Context, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	sequence := s.partner.NextSequence()

	req := &http.Request{
		Method: http.MethodGet,
		URL:    s.partner.StatusURL(),
		Header: map[string][]string{"accept": {"application/json"}},
	}

	cli := &http.Client{
		Timeout: s.cfg.PartnerStatusTimeout,
	}

	res, err := cli.Do(req)
	if err != nil {
		l.Debug("Failed to query partner bridge status",
			zap.Error(err),
			zap.String("bridge_name", s.cfg.Name),
		)
		s.events <- &event.PartnerPollFailure{ // emit event
			Sequence:  sequence,
			Timestamp: time.Now(),
		}
		return
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		l.Warn("Failed to read partner bridge status",
			zap.Error(err),
			zap.String("bridge_name", s.cfg.Name),
		)
		s.events <- &event.PartnerPollFailure{ // emit event
			Sequence:  sequence,
			Timestamp: time.Now(),
		}
		return
	}

	partnerStatus := &types.BridgeStatus{}
	if err := json.Unmarshal(b, partnerStatus); err != nil {
		l.Warn("Failed to parse partner bridge status",
			zap.Error(err),
			zap.String("bridge_name", s.cfg.Name),
		)
		s.events <- &event.PartnerPollFailure{ // emit event
			Sequence:  sequence,
			Timestamp: time.Now(),
		}
		return
	}

	s.events <- &event.PartnerPollSuccess{ // emit event
		Status:    partnerStatus,
		Sequence:  sequence,
		Timestamp: time.Now(),
	}
}
