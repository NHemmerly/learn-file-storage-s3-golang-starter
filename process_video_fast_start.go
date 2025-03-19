package main

import (
	"fmt"
	"os/exec"
)

func processVideoForFastStart(filePath string) (string, error) {
	outPath := fmt.Sprintf("%s%s", filePath, ".processing")
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("could not run command: %w", err)
	}
	return outPath, nil
}
