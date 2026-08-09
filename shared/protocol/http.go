package protocol

type HTTPReq struct {
	ID     string            `json:"id"`
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Header map[string]string `json:"header"`
	Body   []byte            `json:"body"`
}

type HTTPRes struct {
	ID     string            `json:"id"`
	Status int               `json:"status"`
	Header map[string]string `json:"header"`
	Body   []byte            `json:"body"`
}
