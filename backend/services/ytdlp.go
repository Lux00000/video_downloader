package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Format struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Quality string `json:"quality"`
	Ext     string `json:"ext"`
	Size    int64  `json:"size,omitempty"`
}

type VideoInfo struct {
	Platform  Platform `json:"platform"`
	Title     string   `json:"title"`
	Duration  int      `json:"duration"`
	Thumbnail string   `json:"thumbnail"`
	Formats   []Format `json:"formats"`
}

type YtDlpService struct {
	ytdlpPath string
	validator *Validator
}

func NewYtDlpService(ytdlpPath string, validator *Validator) *YtDlpService {
	return &YtDlpService{
		ytdlpPath: ytdlpPath,
		validator: validator,
	}
}

type ytdlpFormat struct {
	FormatID   string  `json:"format_id"`
	Ext        string  `json:"ext"`
	Resolution string  `json:"resolution"`
	VCodec     string  `json:"vcodec"`
	ACodec     string  `json:"acodec"`
	Filesize   int64   `json:"filesize"`
	ABR        float64 `json:"abr"`
	Height     int     `json:"height"`
	FormatNote string  `json:"format_note"`
}

type ytdlpInfo struct {
	Title     string        `json:"title"`
	Duration  float64       `json:"duration"`
	Thumbnail string        `json:"thumbnail"`
	Formats   []ytdlpFormat `json:"formats"`
	Extractor string        `json:"extractor"`
}

func (s *YtDlpService) Analyze(ctx context.Context, url string) (*VideoInfo, error) {
	platform, err := s.validator.ValidateURL(url)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, s.ytdlpPath,
		"--dump-json",
		"--no-download",
		"--no-warnings",
		"--no-playlist",
		url,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute yt-dlp: %w", err)
	}

	var info ytdlpInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	duration := int(info.Duration)

	formats := s.parseFormats(info.Formats)

	return &VideoInfo{
		Platform:  platform,
		Title:     info.Title,
		Duration:  duration,
		Thumbnail: info.Thumbnail,
		Formats:   formats,
	}, nil
}

func (s *YtDlpService) parseFormats(ytFormats []ytdlpFormat) []Format {
	var formats []Format
	seen := make(map[string]bool)

	for _, f := range ytFormats {
		if f.FormatID == "" {
			continue
		}

		var formatType, quality string

		if f.VCodec == "none" && f.ACodec != "none" {
			formatType = "audio"
			if f.ABR > 0 {
				quality = fmt.Sprintf("%.0fkbps", f.ABR)
			} else {
				quality = "audio"
			}
		} else if f.VCodec != "none" {
			formatType = "video"
			// Prefer only mp4 container for better compatibility on Windows
			if f.Ext != "mp4" {
				continue
			}
			if f.Height > 0 {
				quality = fmt.Sprintf("%dp", f.Height)
			} else if f.Resolution != "" && f.Resolution != "audio only" {
				quality = f.Resolution
			} else {
				continue
			}
		} else {
			continue
		}

		key := fmt.Sprintf("%s-%s-%s", formatType, quality, f.Ext)
		if seen[key] {
			continue
		}
		seen[key] = true

		formats = append(formats, Format{
			ID:      f.FormatID,
			Type:    formatType,
			Quality: quality,
			Ext:     f.Ext,
			Size:    f.Filesize,
		})
	}

	return formats
}

// DownloadToFile downloads video to a temp file and returns the file path and filename
func (s *YtDlpService) DownloadToFile(ctx context.Context, url, formatID, tempDir string) (filePath string, filename string, err error) {
	_, err = s.validator.ValidateURL(url)
	if err != nil {
		return "", "", err
	}

	// Generate unique filename prefix
	timestamp := time.Now().UnixNano()
	outputTemplate := filepath.Join(tempDir, fmt.Sprintf("%d_%%(title)s.%%(ext)s", timestamp))

	// Build arguments
	args := []string{
		"-f", formatID,
		"-o", outputTemplate,
		"--no-warnings",
		"--no-playlist",
		"--no-mtime",
	}

	// For merged formats (video+audio), explicitly set output format to mp4
	// This ensures ffmpeg properly merges the streams into a valid container
	if strings.Contains(formatID, "+") {
		args = append(args,
			"--merge-output-format", "mp4",
			"--postprocessor-args", "ffmpeg:-c:v copy -c:a aac -strict experimental",
		)
	}

	args = append(args, url)

	cmd := exec.CommandContext(ctx, s.ytdlpPath, args...)
	cmd.Stderr = os.Stderr // Log errors

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", "", fmt.Errorf("download failed: %s", string(exitErr.Stderr))
		}
		return "", "", fmt.Errorf("download failed: %w", err)
	}

	// Parse output to find downloaded file path
	// yt-dlp prints the destination path
	outputStr := string(output)
	_ = outputStr

	// Find the downloaded file by pattern
	pattern := filepath.Join(tempDir, fmt.Sprintf("%d_*", timestamp))
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", "", fmt.Errorf("could not find downloaded file")
	}

	// Get the first match (should be only one)
	filePath = matches[0]

	// Extract filename without timestamp prefix
	baseName := filepath.Base(filePath)
	// Remove timestamp prefix (format: "1234567890_")
	parts := strings.SplitN(baseName, "_", 2)
	if len(parts) > 1 {
		filename = parts[1]
	} else {
		filename = baseName
	}

	// Verify file exists and has content
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", "", fmt.Errorf("downloaded file not found: %w", err)
	}
	if fileInfo.Size() == 0 {
		os.Remove(filePath)
		return "", "", fmt.Errorf("downloaded file is empty")
	}

	return filePath, filename, nil
}

func (s *YtDlpService) GetBestFormats(formats []Format) []Format {
	var best []Format

	// Find best audio format
	var bestAudio *Format
	for i := range formats {
		f := &formats[i]
		if f.Type == "audio" {
			if bestAudio == nil {
				bestAudio = f
			} else {
				currentBitrate := extractBitrate(f.Quality)
				bestBitrate := extractBitrate(bestAudio.Quality)
				if currentBitrate > bestBitrate {
					bestAudio = f
				}
			}
		}
	}

	// Add best audio option
	if bestAudio != nil {
		best = append(best, Format{
			ID:      bestAudio.ID,
			Type:    "audio",
			Quality: "Лучшее аудио (" + bestAudio.Quality + ")",
			Ext:     bestAudio.Ext,
			Size:    bestAudio.Size,
		})
	}

	// Find best video formats by resolution and create video+audio combos
	resolutions := []int{360, 480, 720, 1080}
	resLabels := map[int]string{360: "360p", 480: "480p", 720: "720p HD", 1080: "1080p Full HD"}

	for _, res := range resolutions {
		for _, f := range formats {
			if f.Type == "video" && strings.HasPrefix(f.Quality, strconv.Itoa(res)) {
				// Video with audio (merged)
				if bestAudio != nil {
					best = append(best, Format{
						ID:      f.ID + "+" + bestAudio.ID,
						Type:    "video",
						Quality: resLabels[res] + " (видео + аудио)",
						Ext:     "mp4",
						Size:    f.Size + bestAudio.Size,
					})
				}
				// Video only (no audio)
				best = append(best, Format{
					ID:      f.ID,
					Type:    "video_only",
					Quality: resLabels[res] + " (только видео)",
					Ext:     f.Ext,
					Size:    f.Size,
				})
				break
			}
		}
	}

	return best
}

// GetFilename returns the filename for a given URL and format without downloading
func (s *YtDlpService) GetFilename(ctx context.Context, url, formatID string) (string, error) {
	// For merged formats, use the base format ID
	baseFormatID := formatID
	if strings.Contains(formatID, "+") {
		parts := strings.Split(formatID, "+")
		baseFormatID = parts[0]
	}

	cmd := exec.CommandContext(ctx, s.ytdlpPath,
		"--get-filename",
		"-f", baseFormatID,
		"-o", "%(title)s.%(ext)s",
		"--no-warnings",
		url,
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get filename: %w", err)
	}

	filename := strings.TrimSpace(string(output))

	// If merging, change extension to mp4
	if strings.Contains(formatID, "+") && !strings.HasSuffix(filename, ".mp4") {
		parts := strings.Split(filename, ".")
		if len(parts) > 1 {
			parts[len(parts)-1] = "mp4"
			filename = strings.Join(parts, ".")
		}
	}

	return filename, nil
}

func extractBitrate(quality string) int {
	quality = strings.TrimSuffix(quality, "kbps")
	if bitrate, err := strconv.Atoi(quality); err == nil {
		return bitrate
	}
	return 0
}
