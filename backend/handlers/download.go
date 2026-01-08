package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"viddown/services"
)

type DownloadHandler struct {
	ytdlp     *services.YtDlpService
	semaphore *services.Semaphore
	logger    *slog.Logger
}

func NewDownloadHandler(ytdlp *services.YtDlpService, semaphore *services.Semaphore, logger *slog.Logger) *DownloadHandler {
	return &DownloadHandler{
		ytdlp:     ytdlp,
		semaphore: semaphore,
		logger:    logger,
	}
}

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	videoURL := r.URL.Query().Get("url")
	formatID := r.URL.Query().Get("format_id")
	formatType := r.URL.Query().Get("type") // "audio", "video", or "video_only"

	if videoURL == "" {
		http.Error(w, `{"error": "URL parameter is required"}`, http.StatusBadRequest)
		return
	}

	// URL decode
	decodedURL, err := url.QueryUnescape(videoURL)
	if err != nil {
		h.logger.Error("Failed to decode URL", "error", err)
		http.Error(w, `{"error": "Invalid URL encoding"}`, http.StatusBadRequest)
		return
	}

	if formatID == "" {
		formatID = "best"
	}

	// Check if this is an audio-only download
	isAudioOnly := formatType == "audio"

	// Try to acquire semaphore (limit concurrent downloads)
	if !h.semaphore.TryAcquire() {
		h.logger.Warn("Too many concurrent downloads", "available", h.semaphore.Available())
		http.Error(w, `{"error": "Server busy. Please try again in a moment."}`, http.StatusServiceUnavailable)
		return
	}
	defer h.semaphore.Release()

	h.logger.Info("Starting download", "url", decodedURL, "format", formatID)

	ctx := r.Context()
	startTime := time.Now()

	// Create temp directory if it doesn't exist
	tempDir := "/tmp/viddown"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		h.logger.Error("Failed to create temp directory", "error", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Download to temp file first (this ensures proper merging for video+audio formats)
	tempFile, filename, err := h.ytdlp.DownloadToFile(ctx, decodedURL, formatID, tempDir, isAudioOnly)
	if err != nil {
		h.logger.Error("Download failed", "url", decodedURL, "error", err, "duration", time.Since(startTime))
		http.Error(w, `{"error": "Download failed"}`, http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile) // Clean up temp file after streaming

	// Open the downloaded file
	file, err := os.Open(tempFile)
	if err != nil {
		h.logger.Error("Failed to open temp file", "file", tempFile, "error", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get file info for Content-Length
	fileInfo, err := file.Stat()
	if err != nil {
		h.logger.Error("Failed to stat temp file", "file", tempFile, "error", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Set headers
	sanitizedFilename := sanitizeFilename(filename)
	encodedFilename := url.PathEscape(filename)

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".mp4"
		filename += ext
		sanitizedFilename += ext
		encodedFilename = url.PathEscape(filename)
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, sanitizedFilename, encodedFilename))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache")

	// Stream the file to response
	written, err := io.Copy(w, file)
	if err != nil {
		h.logger.Error("Failed to stream file", "file", tempFile, "error", err, "written", written)
		return
	}

	h.logger.Info("Download complete", "url", decodedURL, "filename", filename, "size", fileInfo.Size(), "duration", time.Since(startTime))
}

func sanitizeFilename(filename string) string {
	// Remove or replace problematic characters
	replacer := strings.NewReplacer(
		`"`, "'",
		`\`, "_",
		`/`, "_",
		`:`, "-",
		`*`, "_",
		`?`, "_",
		`<`, "_",
		`>`, "_",
		`|`, "_",
	)
	return replacer.Replace(filename)
}
