package auth

import (
	"encoding/json"
	"net/http"
	"relaygo/relay/internal/user"
	"time"
)

type Handler struct {
	userService *user.Service
}

func NewHandler(userService *user.Service) *Handler {
	return &Handler{
		userService: userService,
	}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	u, err := h.userService.Register(
		r.Context(),
		req.Name,
		req.Email,
		req.Password,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(registerResponse{
		ID:        u.ID.String(),
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	u, err := h.userService.Loginserivce(
		r.Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    u.Token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   86400,
	})
	json.NewEncoder(w).Encode(loginResponse{
		Token:     u.Token,
		CreatedAt: u.CreatedAt,
	})

}

func (h *Handler) Test(w http.ResponseWriter, r *http.Request) {
	id, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return
	}
	u, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    u.ID,
		"name":  u.Name,
		"email": u.Email,
	})
}
