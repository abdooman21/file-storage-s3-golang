package main

import "os/exec"

func processVideoForFastStart(filePath string) (string, error) {
	outPath := filePath + ".processing"

	// ffmpeg and the arguments are -i, the input file path, -c, copy, -movflags, faststart, -f, mp4
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outPath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return outPath, nil

}
