package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// EnsureFonts downloads and extracts DejaVu fonts if they don't exist
func ensureFonts() error {
	fontPath := "fonts/DejaVuSans.ttf"
	fontPathBold := "fonts/DejaVuSans-Bold.ttf"

	// Check if fonts already exist and are valid
	if isValidFont(fontPath) && isValidFont(fontPathBold) {
		return nil
	}

	// Create fonts directory
	os.MkdirAll("fonts", 0755)

	// Download the zip archive
	fmt.Println("Downloading DejaVu fonts...")
	zipPath := "/tmp/dejavu.zip"
	err := downloadFile(zipPath,
		"https://sourceforge.net/projects/dejavu/files/dejavu/2.37/dejavu-fonts-ttf-2.37.zip/download")
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	defer os.Remove(zipPath)

	// Extract specific fonts
	err = extractFontsFromZip(zipPath, "fonts/")
	if err != nil {
		return fmt.Errorf("extraction failed: %v", err)
	}

	fmt.Println("✓ Fonts installed successfully")
	return nil
}

func isValidFont(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	// TTF files should be larger than 100KB
	return info.Size() > 100000
}

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractFontsFromZip(zipPath string, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) == "DejaVuSans.ttf" || filepath.Base(f.Name) == "DejaVuSans-Bold.ttf" {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			outFile, err := os.Create(filepath.Join(destDir, filepath.Base(f.Name)))
			if err != nil {
				rc.Close()
				return err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()

			if err != nil {
				return err
			}
		}
	}
	return nil
}
