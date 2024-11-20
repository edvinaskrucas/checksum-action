package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FileChecksum struct {
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
}

func main() {
	rootDir := flag.String("dir", ".", "Root directory to calculate checksums")
	outputFile := flag.String("output", "checksums.json", "Output file to save checksums")
	ignorePaths := flag.String("ignore", "", "Comma-separated list of paths to ignore (relative to root)")

	flag.Parse()

	ignorePatterns := make([]string, 0)

	if *ignorePaths != "" {
		ignorePatterns = strings.Split(*ignorePaths, ",")
	}

	projectDir, err := filepath.Abs(*rootDir)

	if err != nil {
		fmt.Println("Error generating project dir:", err)

		return
	}

	checksums, err := calculateChecksums(projectDir, ignorePatterns)

	if err != nil {
		fmt.Println("Error calculating checksums:", err)

		return
	}

	checksumsFilePath := filepath.Join(projectDir, *outputFile)

	err = saveToFile(checksums, checksumsFilePath)

	if err != nil {
		fmt.Println("Error saving checksums:", err)
	}
}

func generateSHA1Checksum(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)

	if err != nil {
		return "", err
	}

	hash := sha1.Sum(data)

	return hex.EncodeToString(hash[:]), nil
}

func calculateChecksums(rootDir string, ignorePatterns []string) ([]FileChecksum, error) {
	var checksums []FileChecksum

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		ignored, err := isIgnored(path, ignorePatterns, rootDir)

		if err != nil {
			return err
		}

		if ignored {
			if d.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if d.IsDir() {
			return nil
		}

		checksum, err := generateSHA1Checksum(path)

		if err != nil {
			return fmt.Errorf("failed to calculate checksum for %s: %w", path, err)
		}

		relativePath, err := filepath.Rel(rootDir, path)

		if err != nil {
			return err
		}

		checksums = append(checksums, FileChecksum{
			Path:     relativePath,
			Checksum: checksum,
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the directory: %w", err)
	}

	return checksums, nil
}

func isIgnored(path string, ignorePatterns []string, rootDir string) (bool, error) {
	relativePath, err := filepath.Rel(rootDir, path)

	if err != nil {
		return false, err
	}

	for _, pattern := range ignorePatterns {
		if strings.HasPrefix(relativePath, pattern) {
			return true, nil
		}
	}

	return false, nil
}

func saveToFile(checksums []FileChecksum, outputFile string) error {
	outputData, err := json.MarshalIndent(checksums, "", "  ")

	if err != nil {
		return fmt.Errorf("failed to marshal checksums to JSON: %w", err)
	}

	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		return fmt.Errorf("failed to write checksums to file: %w", err)
	}

	return nil
}
