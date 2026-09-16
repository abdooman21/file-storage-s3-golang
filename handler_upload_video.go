package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/abdooman21/file-storage-s3-golang/internal/auth"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No Token found", nil)
		return
	}
	usrID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "user not validated", err)
		return
	}

	bufsize := 1 << 30
	r.Body = http.MaxBytesReader(w, r.Body, int64(bufsize))

	err = r.ParseMultipartForm(1024)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "vid err", err)
		return
	}

	id := r.PathValue("videoID")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "no valid video id", nil)
		return
	}
	viduuid, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "no valid video id", err)
		return
	}
	vidmeta, err := cfg.db.GetVideo(viduuid)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "no valid video id", err)
		return
	}
	if vidmeta.UserID != usrID {
		respondWithError(w, http.StatusUnauthorized, "user can't access an id that's dosent belong to him", nil)
		return
	}

	file, _, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad file data", err)
		return
	}
	defer file.Close()

	mimbytes := make([]byte, 512)
	_, err = file.Read(mimbytes)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad file data", err)
		return
	}

	vidtype := http.DetectContentType(mimbytes)
	if vidtype != "video/mp4" {
		respondWithError(w, http.StatusUnsupportedMediaType, "vid ext isn't allowed", nil)
		return
	}
	// for future support of multiple vid
	// ext, err := mime.ExtensionsByType(vidtype)
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "reading vid error", err)
	// 	return
	// }

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "reading vid error", err)
		return
	}
	tempfile, err := os.CreateTemp("", "tubely_upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "writing vid error", err)
		return
	}
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()
	_, err = io.Copy(tempfile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "writing vid error", err)
		return
	}
	_, err = tempfile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "reading vid error", err)
		return
	}

	seed := make([]byte, 32)
	_, err = rand.Read(seed)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to generate random filename", err)
		return
	}
	aspect, err := getVideoAspectRatio(tempfile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "file corruption", err)
		return
	}

	newpath, err := processVideoForFastStart(tempfile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "file corruption", err)
		return
	}
	defer os.Remove(newpath)
	filename := fmt.Sprintf("%s/%s.mp4", aspect, hex.EncodeToString(seed))
	faststart_vid, err := os.Open(newpath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "file corruption", err)
		return
	}
	defer faststart_vid.Close()

	params := s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &filename,
		Body:        faststart_vid,
		ContentType: &vidtype,
	}
	_, err = cfg.s3client.PutObject(r.Context(), &params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "writing vid error", err)
		return
	}
	//https://<bucket-name>.s3.<region>.amazonaws.com/<key>
	vidurl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, filename)

	vidmeta.VideoURL = &vidurl
	err = cfg.db.UpdateVideo(vidmeta)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "writing vid error", err)
		return
	}
	respondWithJSON(w, http.StatusAccepted, nil)
}
