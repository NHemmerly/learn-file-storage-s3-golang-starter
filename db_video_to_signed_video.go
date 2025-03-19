package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	bucketKey := strings.Split(*video.VideoURL, ",")
	if len(bucketKey) != 2 {
		return database.Video{}, fmt.Errorf("video url invalid")
	}
	signedURL, err := generatePresignedURL(cfg.s3Client, bucketKey[0], bucketKey[1], time.Hour/4)
	if err != nil {
		return database.Video{}, fmt.Errorf("could not create signed URL: %w", err)
	}
	video.VideoURL = &signedURL
	return video, nil
}
