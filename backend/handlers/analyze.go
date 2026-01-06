package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"viddown/services"
)

type AnalyzeHandler struct {
	ytdlp  *services.YtDlpService
	logger *slog.Logger
}

func NewAnalyzeHandler(ytdlp *services.YtDlpService, logger *slog.Logger) *AnalyzeHandler {
	return &AnalyzeHandler{
		ytdlp:  ytdlp,
		logger: logger,
	}
}

type AnalyzeRequest struct {
	URL string `json:"url"`
}

type AnalyzeResponse struct {
	Platform  string            `json:"platform"`
	Title     string            `json:"title"`
	Duration  int               `json:"duration"`
	Thumbnail string            `json:"thumbnail"`
	Formats   []services.Format `json:"formats"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *AnalyzeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "URL is required"})
		return
	}

	h.logger.Info("Analyzing URL", "url", req.URL)

	info, err := h.ytdlp.Analyze(r.Context(), req.URL)
	if err != nil {
		h.logger.Error("Failed to analyze URL", "url", req.URL, "error", err)
		
		w.Header().Set("Content-Type", "application/json")
		
		switch err {
		case services.ErrInvalidURL:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid URL format"})
		case services.ErrUnsupportedURL:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Unsupported platform. Supported: YouTube, Instagram, TikTok"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to analyze video. Please check the URL and try again."})
		}
		return
	}

	// Get simplified formats
	simplifiedFormats := h.ytdlp.GetBestFormats(info.Formats)
	if len(simplifiedFormats) == 0 {
		simplifiedFormats = info.Formats
	}

	response := AnalyzeResponse{
		Platform:  string(info.Platform),
		Title:     info.Title,
		Duration:  info.Duration,
		Thumbnail: info.Thumbnail,
		Formats:   simplifiedFormats,
	}

	h.logger.Info("Analysis complete", "url", req.URL, "title", info.Title, "formats", len(response.Formats))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


