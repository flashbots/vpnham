package httplogger

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/metrics"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

func Middleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID (`base64` to shorten its string representation)
		_uuid := [16]byte(uuid.New())
		httpRequestID := base64.RawStdEncoding.EncodeToString(_uuid[:])

		l := logger.With(
			zap.String("http_request_id", httpRequestID),
			zap.String("log_type", "activity"),
		)
		r = logutils.RequestWithLogger(r, l)

		// Handle panics
		defer func() {
			if msg := recover(); msg != nil {
				w.WriteHeader(http.StatusInternalServerError)
				var method, url string
				if r != nil {
					method = r.Method
					url = r.URL.EscapedPath()
				}
				l.Error("HTTP request handler panicked",
					zap.Any("error", msg),
					zap.String("method", method),
					zap.String("url", url),
				)
				metrics.Errors.Add(r.Context(), 1, otelapi.WithAttributes(
					attribute.String(metrics.LabelErrorScope, metrics.ScopeHTTPMiddleware),
				))
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		// Passing request stats both in-message (for the human reader)
		// as well as inside the structured log (for the machine parser)
		logger.Debug(fmt.Sprintf("%s %s %d", r.Method, r.URL.EscapedPath(), wrapped.Status()),
			zap.Int("duration_ms", int(time.Since(start).Milliseconds())),
			zap.Int("status", wrapped.Status()),
			zap.String("http_request_id", httpRequestID),
			zap.String("log_type", "access"),
			zap.String("method", r.Method),
			zap.String("path", r.URL.EscapedPath()),
			zap.String("user_agent", r.Header.Get("user-agent")),
		)
	})
}
