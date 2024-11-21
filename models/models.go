package models

import (
	"mime"
	"time"
)

// FileType represents the MIME type of an asset file.
type FileType string

// Enum values for FileType
const (
	AudioDescriptionNormal       FileType = "audio/mp3"
	TranscriptSRT                FileType = "text/plain" // no official mimetype for SRT
	SubtitlesVTT                 FileType = "text/vtt"
	TranscriptTimeCoded          FileType = "text/plain"
	ThumbnailJPG                 FileType = "image/jpeg"
	SignLanguagePictureInPicture FileType = "video/mp4"
)

// Asset represents a media asset in the project.
type Asset struct {
	Name     string   `json:"name"`
	FileType FileType `json:"file_type"`
}

// Project represents a project containing multiple assets.
type Project struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Assets []Asset `json:"assets"`
}

// Optionally, a function to lookup file types (using mime if needed)
func GetFileType(extension string) FileType {
	switch extension {
	case "mp3":
		return AudioDescriptionNormal
	case "srt":
		return TranscriptSRT
	case "vtt":
		return SubtitlesVTT
	case "txt":
		return TranscriptTimeCoded
	case "jpeg", "jpg":
		return ThumbnailJPG
	case "mp4":
		return SignLanguagePictureInPicture
	default:
		// Return empty or default type, or fallback to using mime package if needed.
		return FileType(mime.TypeByExtension("." + extension))
	}
}

// Client holds information about each client connection
type Client struct {
	ProjectID   string
	Channel     chan string
	ConnectTime time.Time
}
