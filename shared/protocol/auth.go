package protocol

type AuthReq struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type AuthSuccessRes struct {
	AgentID   string `json:"agent_id"`
	Name      string `json:"name"`
	TunnelURL string `json:"tunnel_url"`
}

type AuthFailedRes struct {
	Reason string `json:"reason"`
}
