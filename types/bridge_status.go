package types

type BridgeStatus struct {
	Name string `json:"name"`

	Active bool `json:"active"`
	Up     bool `json:"up"`

	Role Role `json:"role"`

	Interfaces map[string]*TunnelInterfaceStatus `json:"interfaces"`
}
