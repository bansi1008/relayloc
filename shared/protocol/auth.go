package protocol

type AuthReq struct {
	Email     string `json:"email"`
	AgentName string `json:"agent_name"`
	AccessID  string `json:"access_id"`
	PublicKey string `json:"public_key"`
}

type AuthSuccessRes struct {
	AgentID   string `json:"agent_id"`
	Name      string `json:"name"`
	TunnelURL string `json:"tunnel_url"`
}

type AuthFailedRes struct {
	Reason string `json:"reason"`
}

type ChallengeReq struct {
	AgentID string `json:"agent_id"`
}

type ChallengeRes struct {
	Nonce string `json:"nonce"`
}

type ChallengeVerify struct {
	AgentID   string `json:"agent_id"`
	Signature string `json:"signature"`
}

