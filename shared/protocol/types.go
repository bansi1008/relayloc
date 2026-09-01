package protocol

type FType string

const (
	RegisterAgent FType = "REGISTER_AGENT"
	Registered    FType = "REGISTERED_AGENT"

	AuthenticateAgent    FType = "AUTHENTICATE_AGENT"
	ChallengeReqFrame    FType = "CHALLENGE_REQ"
	ChallengeResFrame    FType = "CHALLENGE_RES"
	ChallengeVerifyFrame FType = "CHALLENGE_VERIFY"
	AuthSuccess          FType = "AUTH_SUCCESS"
	AuthFailed           FType = "AUTH_FAILED"

	Ping FType = "PING"
	Pong FType = "PONG"
)
const (
	HTTPReqFrame = "HTTP_REQUEST"
	HTTPResFrame = "HTTP_RESPONSE"
)

