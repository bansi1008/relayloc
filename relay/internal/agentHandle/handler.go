package agenthandle

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"relaygo/relay/internal/agent"
	"relaygo/relay/internal/auth"
)

type Handler struct {
	agentservice *agent.Service
}

func NewHandler(agentservice *agent.Service) *Handler {
	return &Handler{
		agentservice: agentservice,
	}
}

type registeragent struct {
	Name string `json:"name"`
}
type registerresagent struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Token     string `json:"token"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return

	}
	var req registeragent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("name %s", req.Name)

	_, err := h.agentservice.GetAgentByName(r.Context(), req.Name, id)
	if err == nil {
		http.Error(w, "agent name already exists, please use a different name", http.StatusBadRequest)
		return
	}
	tok, err := Generate()
	if err != nil {
		return
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(tok),
		bcrypt.DefaultCost,
	)

	u, err := h.agentservice.NewAgent(r.Context(), req.Name, id, string(hash))

	if err != nil {
		log.Printf("err: %v", err)
		//http.Error(w, "something went wrong please try again", http.StatusNotFound)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(registerresagent{
		ID:    u.ID.String(),
		Name:  u.Name,
		Token: tok,
		//CreatedAt: u.CreatedAt.Format(time.RFC3339),
	})

}

func (h *Handler) GetAgentByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	agentID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid agent id", http.StatusBadRequest)
		return
	}

	a, err := h.agentservice.GetAgentByID(r.Context(), agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

func (h *Handler) GetAgentsByUserID(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return
	}

	agents, err := h.agentservice.GetAgentsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (h *Handler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	agentID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid agent id", http.StatusBadRequest)
		return
	}

	if err := h.agentservice.DeleteAgent(r.Context(), agentID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func Generate() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
