package server

import ("net/http"  	
        "relaygo/internal/tunnel"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	reg := tunnel.NewRegistry()
	wsServer := NewWSServer(reg)
	httpServer := NewHTTPServer(reg)



	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Printf("Received health check from %s\n", r.RemoteAddr)
		JSON(w, http.StatusOK, map[string]string{
        "status": "ok",
    })
	})
   mux.HandleFunc("/connect", wsServer.HandleConnect)
	mux.HandleFunc("/t/", httpServer.HandleProxy)
	return mux
}
