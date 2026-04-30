package hp

import (
	"encoding/json"
	"ming/internal/service"
	"net/http"
)

type registerReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Sex      int    `json:"sex"`
}

type registerResp struct {
	ID int64 `json:"id"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Password == "" {
		http.Error(w, "name or password is empty", http.StatusBadRequest)
		return
	}
	if req.Sex < 0 {
		http.Error(w, "sex is invalid", http.StatusBadRequest)
		return
	}

	id, err := service.Register(req.Name, req.Password, req.Sex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(registerResp{ID: id}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
