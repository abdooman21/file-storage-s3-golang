package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

type aspect struct {
	Stream []nested_aspect `json:"streams"`
}
type nested_aspect struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	//-v error -print_format json -show_streams samples/boots-video-horizontal.mp4
	// -v, error, -print_format, json, -show_streams
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	buf := bytes.Buffer{}
	cmd.Stdout = &buf

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	js := aspect{}
	err = json.Unmarshal(buf.Bytes(), &js)
	if err != nil {
		return "", err
	}
	hei := js.Stream[0].Height
	wei := js.Stream[0].Width
	ratio := float64(wei) / float64(hei)

	if math.Abs(ratio-16.0/9.0) < 0.01 {
		//landscape (16:9 aspect ratio)  portrait ((9:16))
		return "landscape", nil
	}
	if math.Abs(ratio-9.0/16.0) < 0.01 {
		return "portrait", nil
	}
	return "other", nil
}
