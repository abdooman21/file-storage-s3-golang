package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/abdooman21/file-storage-s3-golang/internal/auth"
	"github.com/google/uuid"
)

const maxMemory = 10 << 20

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
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
		respondWithError(w, http.StatusBadRequest, "Failed to parse multipart form", err)
		return
	}
	img, parts, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to get thumbnail image", err)
		return
	}
	defer img.Close()
	ctype := parts.Header.Get("Content-Type")
	midtype, _, err := mime.ParseMediaType(ctype)
	if err != nil || (midtype != "image/jpeg" && midtype != "image/png") {
		respondWithError(w, http.StatusUnsupportedMediaType, "not accepted type", err)
		return
	}
	ext, err := mime.ExtensionsByType(ctype)
	if err != nil || len(ext) == 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid thumbnail image type", err)
		return
	}

	// byt, err := io.ReadAll(img)
	// if err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "failed to parse image", err)
	// 	return
	// }

	vid, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch video data", err)
		return
	}
	if vid.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You don't have permission to update this video", nil)
		return
	}

	// enc_img := base64.StdEncoding.EncodeToString(byt)
	// //data:image/png;base64,iVBORw0KGgoAAA...
	// url := fmt.Sprintf("data:%s;base64,%s", ctype, enc_img)

	path := filepath.Join(cfg.assetsRoot, vid.ID.String()+ext[0])
	file, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save thumbnail", err)
		return
	}
	defer file.Close()
	_, err = io.Copy(file, img)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save thumbnail", err)
		return
	}
	urlPath := fmt.Sprintf("http://localhost:%s/assets/%s%s", cfg.port, vid.ID.String(), ext[0])
	vid.ThumbnailURL = &urlPath
	err = cfg.db.UpdateVideo(vid)
	if err != nil {
		respondWithError(w, 500, "Failed to update video data", err)
		return
	}

	respondWithJSON(w, http.StatusOK, vid)
}
