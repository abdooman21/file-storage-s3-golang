package main

import (
	"bytes"
	"os"
	"testing"
)

func TestFaststart(t *testing.T) {
	vidPath := "samples/boots-video-vertical.mp4"
	outPath, err := processVideoForFastStart(vidPath)

	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if outPath != "samples/boots-video-vertical.mp4.processing" {
		t.Log(err)
		t.Fail()
	}
	defer os.Remove(outPath)
	// Source - https://stackoverflow.com/a/78673947
	// Posted by Nathan Smith
	// Retrieved 2026-09-17, License - CC BY-SA 4.0

	/////////////////////////////////////////////////////////////////////////////////
	// Fast start is enabled if "moov" is before "mdat" in first 4096 bytes fo file
	// https://trac.ffmpeg.org/wiki/HowToCheckIfFaststartIsEnabledForPlayback
	/////////////////////////////////////////////////////////////////////////////////

	fs, err := os.Open(outPath)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	buff := make([]byte, 4096)
	fs.Read(buff)
	defer fs.Close()

	// moovPattern := []byte{'m', 'o', 'o', 'v', 0x00}  equivilant
	var moovPattern = []byte{0x6d, 0x6f, 0x6f, 0x76, 0x00}
	moovIndex := bytes.Index(buff, moovPattern)
	// mdatPattern := []byte{'m', 'd', 'a', 't', 0x00} equivilant
	var mdatPattern = []byte{0x6d, 0x64, 0x61, 0x74, 0x00}
	mdatIndex := bytes.Index(buff, mdatPattern)

	if moovIndex < 0 || (mdatIndex != -1 && moovIndex > mdatIndex) {
		// Fast start is enabled
		t.Fail()
	}

	vid, err := os.Open(vidPath)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	vid.Read(buff)
	defer vid.Close()
	moovIndex = bytes.Index(buff, moovPattern)
	mdatIndex = bytes.Index(buff, mdatPattern)
	//fast start should be diabled here
	if moovIndex >= 0 && (mdatIndex == -1 || moovIndex < mdatIndex) {
		t.Fail()
	}

}
