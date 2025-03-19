package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)

	videoID, err := uuid.Parse(r.PathValue("videoID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	vidMeta, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "video not found", err)
		return
	}
	if userID != vidMeta.UserID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	video, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
	}
	defer video.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusForbidden, "incorrect filetype", err)
		return
	}

	videoFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create temp file", err)
		return
	}
	defer os.Remove(videoFile.Name())
	defer videoFile.Close()

	_, err = io.Copy(videoFile, video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not copy video to temp", err)
		return
	}
	processedFileName, err := processVideoForFastStart(videoFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
		return
	}
	processedFile, err := os.Open(processedFileName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not retrieve video", err)
		return
	}
	videoFile.Seek(0, io.SeekStart)
	aspectRatio, err := getVideoAspectRatio(processedFileName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get video aspect ratio", err)
		return
	}
	var prefix string
	switch aspectRatio {
	case "16:9":
		prefix = "landscape/"
	case "9:16":
		prefix = "portrait/"
	default:
		prefix = "other/"
	}

	name := make([]byte, 32)
	rand.Read(name)
	keyName := fmt.Sprintf("%s%s", prefix, hex.EncodeToString(name)+".mp4")

	input := &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &keyName,
		Body:        processedFile,
		ContentType: &mediaType,
	}
	cfg.s3Client.PutObject(r.Context(), input)
	//newVidUrl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, keyName)
	//newVidUrl := fmt.Sprintf("%s,%s", cfg.s3Bucket, keyName)
	newVidUrl := fmt.Sprintf("%s/%s", cfg.s3CfDistribution, keyName)
	vidMeta.VideoURL = &newVidUrl
	if err := cfg.db.UpdateVideo(vidMeta); err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
		return
	}
	/*
		vidMeta, err = cfg.dbVideoToSignedVideo(vidMeta)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "could not update vidURL", err)
			return
		}
	*/
	respondWithJSON(w, http.StatusOK, vidMeta.VideoURL)
}
