package engines

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

func RegModule(regValue string, regKey string, constraints string) []ScanResult {
	var hits []ScanResult

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	targetPath := constraints
	if targetPath == "" {
		targetPath = regKey
	}

	var rootKey registry.Key
	var subKeyPath string

	upperPath := strings.ToUpper(targetPath)
	if strings.HasPrefix(upperPath, "HKLM") || strings.HasPrefix(upperPath, "HKEY_LOCAL_MACHINE") {
		rootKey = registry.LOCAL_MACHINE
		if strings.HasPrefix(upperPath, "HKLM") {
			subKeyPath = strings.TrimPrefix(targetPath, "HKLM\\")
		} else {
			subKeyPath = strings.TrimPrefix(targetPath, "HKEY_LOCAL_MACHINE\\")
		}
	} else if strings.HasPrefix(upperPath, "HKCU") || strings.HasPrefix(upperPath, "HKEY_CURRENT_USER") {
		rootKey = registry.CURRENT_USER
		if strings.HasPrefix(upperPath, "HKCU") {
			subKeyPath = strings.TrimPrefix(targetPath, "HKCU\\")
		} else {
			subKeyPath = strings.TrimPrefix(targetPath, "HKEY_CURRENT_USER\\")
		}
	} else {

		rootKey = registry.LOCAL_MACHINE
		subKeyPath = targetPath
	}

	k, err := registry.OpenKey(rootKey, subKeyPath, registry.READ)
	if err != nil {

		hits = append(hits, ScanResult{
			Hostname:  hostname,
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "registry_key",
			Value:     targetPath,
			Status:    "CLEAN",
			Details:   "Key not found",
		})
		return hits
	}
	defer k.Close()

	if regValue != "" {

		_, _, err := k.GetStringValue(regValue)
		if err == nil {
			fmt.Printf("[!] HIT: Registry Value Match on %s\\%s\n", targetPath, regValue)
			hits = append(hits, ScanResult{
				Hostname:  hostname,
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "registry_value",
				Value:     regValue,
				Status:    "DIRTY",
				Details:   fmt.Sprintf("Key: %s", targetPath),
			})
		} else {
			hits = append(hits, ScanResult{
				Hostname:  hostname,
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "registry_value",
				Value:     regValue,
				Status:    "CLEAN",
				Details:   "Value not found",
			})
		}
	} else {

		fmt.Printf("[!] HIT: Registry Key Match on %s\n", targetPath)
		hits = append(hits, ScanResult{
			Hostname:  hostname,
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "registry_key",
			Value:     targetPath,
			Status:    "DIRTY",
			Details:   "Key exists",
		})
	}

	return hits
}
