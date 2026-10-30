package converter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func ConvertToJPG(ctx context.Context, data []byte, filename string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "schedule-bot-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	docPath := filepath.Join(tmpDir, filename)

	if err := os.WriteFile(docPath, data, 0644); err != nil {
		return "", err
	}

	// DOC → PDF
	cmd := exec.CommandContext(
		ctx,
		"libreoffice",
		"-env:UserInstallation=file://"+filepath.Join(tmpDir, "lo-profile"),
		"--headless",
		"--convert-to", "pdf",
		"--outdir", tmpDir,
		docPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("libreoffice: %w: %s", err, output)
	}

	pdfPath := strings.TrimSuffix(docPath, filepath.Ext(docPath)) + ".pdf"

	// PDF → JPG
	jpgPrefix := filepath.Join(tmpDir, "schedule")

	cmd = exec.CommandContext(
		ctx,
		"pdftoppm",
		"-jpeg",
		"-r", "150",
		pdfPath,
		jpgPrefix,
	)

	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pdftoppm: %w: %s", err, output)
	}

	jpgPath := jpgPrefix + "-1.jpg"

	if _, err := os.Stat(jpgPath); err != nil {
		return "", fmt.Errorf("jpg not created: %w", err)
	}

	result, err := os.ReadFile(jpgPath)
	if err != nil {
		return "", err
	}

	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	resultPath := filepath.Join(os.TempDir(), baseName+".jpg")

	if err := os.WriteFile(resultPath, result, 0644); err != nil {
		return "", err
	}

	return resultPath, nil
}
