package hp

import (
	"encoding/json"
	"ming/internal/service"
	"net/http"
)

type loginReq struct {
	ID       int64  `json:"id"`
	Password string `json:"password"`
}

type loginResp struct {
	Token string `json:"token"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if req.ID <= 0 || req.Password == "" {
		http.Error(w, "user_id or password is invalid", http.StatusBadRequest)
		return
	}

	token, err := service.Login(req.ID, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := loginResp{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
