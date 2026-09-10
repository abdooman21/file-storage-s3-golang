package main

import (
	"testing"
)

func TestAspect_ration(t *testing.T) {

	filepath := "samples/boots-video-horizontal.mp4"
	result, err := getVideoAspectRatio(filepath)
	if err != nil {
		t.Error(err)
	}

	if result != "landscape" {
		t.Fail()
	}
	t.Log(result)
	file2path := "samples/boots-video-vertical.mp4"
	res, err := getVideoAspectRatio(file2path)
	if err != nil {
		t.Error(err)
	}
	if res != "portrait" {
		t.Errorf("expected portrait, got %s", res)
	}
	t.Log(res)
}
