package handlers

import (
	"encoding/json"
	"net/http"

	"viddown/config"
)

type ConfigHandler struct {
	cfg *config.Config
}

func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{cfg: cfg}
}

type ConfigResponse struct {
	AuthRequired  bool     `json:"authRequired"`
	MaxConcurrent int      `json:"maxConcurrent"`
	Platforms     []string `json:"platforms"`
}

func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := ConfigResponse{
		AuthRequired:  h.cfg.AuthRequired,
		MaxConcurrent: h.cfg.MaxConcurrent,
		Platforms:     []string{"youtube", "instagram", "tiktok"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


