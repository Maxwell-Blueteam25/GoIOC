package engines

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func FileModule(fileName string, searchPath string) []ScanResult {
	var hits []ScanResult

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	found := false
	fileName = strings.ToLower(fileName)

	if searchPath != "" {
		entries, err := os.ReadDir(searchPath)
		if err == nil {
			for _, d := range entries {
				if d.IsDir() {
					continue
				}
				if strings.ToLower(d.Name()) == fileName {
					found = true
					fullPath := filepath.Join(searchPath, d.Name())
					fmt.Printf("[!] HIT: File Match on %s\n", fileName)
					hits = append(hits, ScanResult{
						Hostname:  hostname,
						Timestamp: time.Now().Format(time.RFC3339),
						Type:      "file_name",
						Value:     fileName,
						Status:    "DIRTY",
						Details:   fmt.Sprintf("Path: %s", fullPath),
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
			if strings.ToLower(d.Name()) == fileName {
				found = true
				fmt.Printf("[!] HIT: File Match on %s\n", fileName)
				hits = append(hits, ScanResult{
					Hostname:  hostname,
					Timestamp: time.Now().Format(time.RFC3339),
					Type:      "file_name",
					Value:     fileName,
					Status:    "DIRTY",
					Details:   fmt.Sprintf("Path: %s", path),
				})
			}
			return nil
		})
	}

	if !found {
		hits = append(hits, ScanResult{
			Hostname:  hostname,
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "file_name",
			Value:     fileName,
			Status:    "CLEAN",
			Details:   "N/A",
		})
	}

	return hits
}
