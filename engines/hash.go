package engines

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func HashModule(hashValue string, searchPath string) []ScanResult {
	var hits []ScanResult

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	hashValue = strings.ToLower(hashValue)
	found := false

	if searchPath != "" {
		entries, err := os.ReadDir(searchPath)
		if err == nil {
			for _, d := range entries {
				if d.IsDir() {
					continue
				}

				fullPath := filepath.Join(searchPath, d.Name())
				file, err := os.Open(fullPath)
				if err != nil {
					continue
				}

				hasher := sha256.New()
				_, errCopy := io.Copy(hasher, file)
				file.Close()

				if errCopy != nil {
					continue
				}

				fileHash := hex.EncodeToString(hasher.Sum(nil))

				if fileHash == hashValue {
					found = true
					fmt.Printf("[!] HIT: Hash Match on %s\n", fullPath)
					hits = append(hits, ScanResult{
						Hostname:  hostname,
						Timestamp: time.Now().Format(time.RFC3339),
						Type:      "file_hash",
						Value:     hashValue,
						Status:    "DIRTY",
						Details:   fullPath,
					})
				}
			}
		}
	} else {
		searchPath = "C:\\"
		filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer file.Close()

			hasher := sha256.New()
			if _, err := io.Copy(hasher, file); err != nil {
				return nil
			}

			fileHash := hex.EncodeToString(hasher.Sum(nil))

			if fileHash == hashValue {
				found = true
				fmt.Printf("[!] HIT: Hash Match on %s\n", path)
				hits = append(hits, ScanResult{
					Hostname:  hostname,
					Timestamp: time.Now().Format(time.RFC3339),
					Type:      "file_hash",
					Value:     hashValue,
					Status:    "DIRTY",
					Details:   path,
				})
			}
			return nil
		})
	}

	if !found {
		hits = append(hits, ScanResult{
			Hostname:  hostname,
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "file_hash",
			Value:     hashValue,
			Status:    "CLEAN",
			Details:   "N/A",
		})
	}

	return hits
}
