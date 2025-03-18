package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func getVideoAspectRatio(filepath string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)
	cmd.Stdout = &out
	cmd.Run()
	var output stdout
	if err := json.Unmarshal(out.Bytes(), &output); err != nil {
		return "", fmt.Errorf("could not unmarshal stdout: %w", err)
	}
	ratio := float64(output.Streams[0].Width) / float64(output.Streams[0].Height)
	fmt.Printf("%.2f\n", ratio)
	if ratio >= 1.75 && ratio <= 1.8 {
		return "16:9", nil
	} else if ratio >= 0.55 && ratio <= 0.57 {
		return "9:16", nil
	} else {
		return "other", nil
	}
}
