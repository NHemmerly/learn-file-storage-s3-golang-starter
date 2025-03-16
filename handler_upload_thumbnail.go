package main

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	const maxMemory = 10 << 20
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
		return
	}
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
		return
	}
	userId, err := auth.ValidateJWT(bearer, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
		return
	}
	meta, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "video not found", err)
		return
	}
	if userId != meta.UserID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	mimeType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	fmt.Print(mimeType)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		respondWithError(w, http.StatusForbidden, "incorrect filetype", err)
		return
	}
	extension := strings.Split(header.Filename, ".")
	newThumbnailUrl := fmt.Sprintf("http://localhost:%s/assets/%v.%s", cfg.port, videoID, extension[len(extension)-1])
	meta.ThumbnailURL = &newThumbnailUrl
	filepath := filepath.Join(cfg.assetsRoot, videoIDString+"."+extension[len(extension)-1])
	newFile, err := os.Create(filepath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
		return
	}
	if _, err = io.Copy(newFile, bytes.NewReader(data)); err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
		return
	}

	if err := cfg.db.UpdateVideo(meta); err != nil {
		respondWithError(w, http.StatusInternalServerError, "server error", err)
		return
	}

	respondWithJSON(w, http.StatusOK, database.Video{
		ID:           meta.ID,
		CreatedAt:    meta.CreatedAt,
		UpdatedAt:    meta.UpdatedAt,
		ThumbnailURL: meta.ThumbnailURL,
		VideoURL:     meta.VideoURL,
		CreateVideoParams: database.CreateVideoParams{
			Title:       meta.Title,
			Description: meta.Description,
			UserID:      meta.UserID},
	})

}
