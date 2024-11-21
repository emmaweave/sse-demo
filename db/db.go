package db

import "sse-demo/models"

func GetProjectFromDB() models.Project {
	return models.Project{
		ID:   "12",
		Name: "My video on guide dogs.",
		Assets: []models.Asset{
			{
				Name:     "Audio Description",
				FileType: models.AudioDescriptionNormal,
			},
			{
				Name:     "Transcript (SRT)",
				FileType: models.TranscriptSRT,
			},
			{
				Name:     "Subtitles (VTT)",
				FileType: models.SubtitlesVTT,
			},
			{
				Name:     "Transcript (Time-Coded)",
				FileType: models.TranscriptTimeCoded,
			},
			{
				Name:     "Thumbnail Image",
				FileType: models.ThumbnailJPG,
			},
			{
				Name:     "Sign Language (Picture in Picture)",
				FileType: models.SignLanguagePictureInPicture,
			},
		},
	}
}
