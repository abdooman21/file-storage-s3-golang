package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/abdooman21/file-storage-s3-golang/internal/auth"
	"github.com/google/uuid"
)

const maxMemory = 10 << 20

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header) // switch later to a middleware
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

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to parse data", err)
		return
	}
	img, parts, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to parse image", err)
		return
	}
	ctype := parts.Header.Get("Content-Type")

	byt, err := io.ReadAll(img)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to parse image", err)
		return
	}

	vid, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, 500, "try again cant fetch data", err)
		return
	}
	if vid.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "out of reach", err)
		return
	}
	enc_img := base64.StdEncoding.EncodeToString(byt)

	//data:image/png;base64,iVBORw0KGgoAAA...
	url := fmt.Sprintf("data:%s;base64,%s", ctype, enc_img)
	vid.ThumbnailURL = &url
	err = cfg.db.UpdateVideo(vid)
	if err != nil {
		respondWithError(w, 500, "try again cant fetch data", err)
		return
	}

	respondWithJSON(w, http.StatusOK, vid)
}
