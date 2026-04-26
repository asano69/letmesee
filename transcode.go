package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var webmCache sync.Map

func TranscodeToWebM(mpeg1Data []byte) ([]byte, error) {
	key := sha256.Sum256(mpeg1Data)
	if v, ok := webmCache.Load(key); ok {
		return v.([]byte), nil
	}

	inFile, err := os.CreateTemp("", "lms-in-*.mpg")
	if err != nil {
		return nil, fmt.Errorf("create input temp file: %w", err)
	}
	defer os.Remove(inFile.Name())
	if _, err := inFile.Write(mpeg1Data); err != nil {
		inFile.Close()
		return nil, fmt.Errorf("write input temp file: %w", err)
	}
	inFile.Close()

	outFile, err := os.CreateTemp("", "lms-out-*.webm")
	if err != nil {
		return nil, fmt.Errorf("create output temp file: %w", err)
	}
	outPath := outFile.Name()
	outFile.Close()
	defer os.Remove(outPath)

	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", inFile.Name(),
		"-c:v", "libvpx",
		"-crf", "10",
		"-b:v", "0",
		"-f", "webm",
		outPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w\n%s", err, out)
	}

	webm, err := os.ReadFile(outPath)
	if err != nil {
		return nil, err
	}
	webmCache.Store(key, webm)
	return webm, nil
}
