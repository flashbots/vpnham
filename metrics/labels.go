package metrics

const (
	LabelBridge = "bridge"
	LabelTunnel = "tunnel"

	LabelProbeSrc = "probe_location_src"
	LabelProbeDst = "probe_location_dst"

	LabelErrorScope = "scope"
)

const (
	ScopeAWS            = "aws"
	ScopeHTTPMiddleware = "http_middleware"
	ScopeInternalLogic  = "internal_logic"
	ScopePartnerPolling = "partner_polling"
	ScopePeerProbing    = "peer_probing"
	ScopeStatusListener = "status_listener"
	ScopeSystem         = "system"
)
