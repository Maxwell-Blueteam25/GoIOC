package engines

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// Defining the struct exactly as requested
type ScanResult struct {
	Hostname  string
	Timestamp string
	Type      string
	Value     string
	Status    string
	Details   string // Added this so we can store the CMD/PID
}

func ProcessModule(processValue string) []ScanResult {
	var hits []ScanResult

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	// List all processes
	processList, err := process.Processes()
	if err != nil {
		log.Printf("Error Retrieving processes: %v", err)
		return nil
	}

	targetName := strings.ToLower(processValue)
	found := false

	// Loop over all processes
	for _, p := range processList {
		name, err := p.Name()
		if err != nil {
			continue
		}

		if strings.ToLower(name) == targetName {
			found = true
			pid := p.Pid
			cmdLine, _ := p.Cmdline()

			fmt.Printf("[!] HIT: Process Match on %s\n", processValue)

			hits = append(hits, ScanResult{
				Hostname:  hostname,
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "process_name", // Normalized to match your JSON config type
				Value:     processValue,
				Status:    "DIRTY",
				Details:   fmt.Sprintf("PID: %d | CMD: %s", pid, cmdLine),
			})
		}
	}

	// This check MUST happen after the loop finishes
	if !found {
		hits = append(hits, ScanResult{
			Hostname:  hostname,
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "process_name",
			Value:     processValue,
			Status:    "CLEAN",
			Details:   "N/A",
		})
	}

	return hits
}
