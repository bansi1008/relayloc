package protocol

type FType string

const (
	RegisterAgent FType = "REGISTER_AGENT"
	Registered    FType = "REGISTERED_AGENT"

	Ping FType = "PING"
	Pong FType = "PONG"
)
