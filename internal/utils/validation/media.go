package validation

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// Supported file formats
var (
	SupportedImageFormats = []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
	SupportedVideoFormats = []string{".mp4", ".mov", ".avi", ".webm", ".mkv"}
	SupportedAudioFormats = []string{".mp3", ".wav", ".ogg", ".m4a", ".flac"}
	SupportedDocFormats   = []string{".pdf", ".doc", ".docx", ".txt", ".rtf"}
)

// MaxFileSize limits for different file types (in bytes)
const (
	MaxImageSize = 5 * 1024 * 1024   // 5MB
	MaxVideoSize = 100 * 1024 * 1024 // 100MB
	MaxAudioSize = 20 * 1024 * 1024  // 20MB
	MaxDocSize   = 10 * 1024 * 1024  // 10MB
	MaxFileSize  = 200 * 1024 * 1024 // 200MB general limit
)

// IsValidFileExtension checks if a file has a valid extension
func IsValidFileExtension(filename string, allowedExtensions []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

// IsValidImageFile checks if a file is a valid image file
func IsValidImageFile(file *multipart.FileHeader) bool {
	// Check file size
	if file.Size > MaxImageSize {
		return false
	}

	// Check file extension
	if !IsValidFileExtension(file.Filename, SupportedImageFormats) {
		return false
	}

	// Check file content (MIME type)
	f, err := file.Open()
	if err != nil {
		return false
	}
	defer f.Close()

	// Read a small portion of the file to detect MIME type
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Reset file cursor
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return false
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)
	return strings.HasPrefix(contentType, "image/")
}

// IsValidVideoFile checks if a file is a valid video file
func IsValidVideoFile(file *multipart.FileHeader) bool {
	// Check file size
	if file.Size > MaxVideoSize {
		return false
	}

	// Check file extension
	if !IsValidFileExtension(file.Filename, SupportedVideoFormats) {
		return false
	}

	// Check file content (MIME type)
	f, err := file.Open()
	if err != nil {
		return false
	}
	defer f.Close()

	// Read a small portion of the file to detect MIME type
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Reset file cursor
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return false
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)
	return strings.HasPrefix(contentType, "video/")
}

// IsValidAudioFile checks if a file is a valid audio file
func IsValidAudioFile(file *multipart.FileHeader) bool {
	// Check file size
	if file.Size > MaxAudioSize {
		return false
	}

	// Check file extension
	if !IsValidFileExtension(file.Filename, SupportedAudioFormats) {
		return false
	}

	// Check file content (MIME type)
	f, err := file.Open()
	if err != nil {
		return false
	}
	defer f.Close()

	// Read a small portion of the file to detect MIME type
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Reset file cursor
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return false
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)
	return strings.HasPrefix(contentType, "audio/")
}

// IsValidMediaFile checks if a file is a valid media file (image, video, or audio)
func IsValidMediaFile(file *multipart.FileHeader) bool {
	return IsValidImageFile(file) || IsValidVideoFile(file) || IsValidAudioFile(file)
}

// GetFileType returns the type of the file based on its extension
func GetFileType(filename string) string {
	strings.ToLower(filepath.Ext(filename))

	if IsValidFileExtension(filename, SupportedImageFormats) {
		return "image"
	} else if IsValidFileExtension(filename, SupportedVideoFormats) {
		return "video"
	} else if IsValidFileExtension(filename, SupportedAudioFormats) {
		return "audio"
	} else if IsValidFileExtension(filename, SupportedDocFormats) {
		return "document"
	}

	return "unknown"
}

// IsValidImageDimensions checks if an image has valid dimensions
func IsValidImageDimensions(width, height int) bool {
	// Minimum dimensions
	minWidth, minHeight := 10, 10

	// Maximum dimensions
	maxWidth, maxHeight := 8192, 8192

	return width >= minWidth && width <= maxWidth && height >= minHeight && height <= maxHeight
}

// HasExifData checks if an image has EXIF data
func HasExifData(data []byte) bool {
	// Check for EXIF marker
	exifMarker := []byte{0xFF, 0xE1}
	return bytes.Contains(data, exifMarker)
}
