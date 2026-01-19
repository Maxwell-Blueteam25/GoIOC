package main

import (
	"bluewave-sweeper/collections"
	"bluewave-sweeper/engines"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xeipuuv/gojsonschema"
)

type SweepConfig struct {
	Sweeps []SweepItem `json:"sweeps"`
}

type SweepItem struct {
	Type        string      `json:"type"`
	Value       string      `json:"value"`
	Constraints Constraints `json:"constraints,omitempty"`
}

type Constraints struct {
	Path string `json:"path,omitempty"`
}

type FinalReport struct {
	Metadata  Metadata             `json:"metadata"`
	Findings  []engines.ScanResult `json:"ioc_matches"`
	Telemetry *CollectionData      `json:"collections,omitempty"`
}

type Metadata struct {
	Hostname  string `json:"hostname"`
	Timestamp string `json:"timestamp"`
	ScanID    string `json:"scan_id"`
}

type CollectionData struct {
	Processes []collections.ProcessData `json:"processes"`
	Network   []collections.NetworkData `json:"network"`
	Services  []collections.ServiceData `json:"services"`
	Registry  []collections.RegData     `json:"registry"`
}

func ParseConfig(filePath string) (*SweepConfig, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config SweepConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func InvokeSweep(inputJsonPath string, reportMode, smbPath, sasUrl string, enableCollections bool) {
	// Debug Print to confirm logic
	if enableCollections {
		fmt.Println("[+] COLLECTION ENGINE: ENABLED")
	} else {
		fmt.Println("[-] COLLECTION ENGINE: DISABLED (Flag not detected)")
	}

	schemaPath := "sweeper-schema.json"

	absSchema, _ := filepath.Abs(schemaPath)
	absDoc, _ := filepath.Abs(inputJsonPath)
	schemaURI := "file:///" + filepath.ToSlash(absSchema)
	docURI := "file:///" + filepath.ToSlash(absDoc)

	schemaLoader := gojsonschema.NewReferenceLoader(schemaURI)
	documentLoader := gojsonschema.NewReferenceLoader(docURI)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		log.Fatalf("Validation Engine Error: %s", err)
	}

	if result.Valid() {
		fmt.Println("Config is Valid. Ready For Deployment.")
	} else {
		fmt.Println("The config is not valid. See errors:")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return
	}

	config, err := ParseConfig(inputJsonPath)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	var findings []engines.ScanResult

	for _, item := range config.Sweeps {
		fmt.Printf("Deploying Scanner -> Type: %s | Target: %s\n", item.Type, item.Value)

		var currentFindings []engines.ScanResult

		switch item.Type {
		case "file_hash":
			currentFindings = engines.HashModule(item.Value, item.Constraints.Path)
		case "registry_value":
			currentFindings = engines.RegModule(item.Value, "", item.Constraints.Path)
		case "registry_key":
			currentFindings = engines.RegModule("", item.Value, "")
		case "process_name":
			currentFindings = engines.ProcessModule(item.Value)
		case "file_name":
			currentFindings = engines.FileModule(item.Value, item.Constraints.Path)
		}

		if len(currentFindings) > 0 {
			findings = append(findings, currentFindings...)
		}
	}

	hostname, _ := os.Hostname()
	timestamp := time.Now().Format("20060102_150405")
	isoTime := time.Now().Format(time.RFC3339)

	finalReport := FinalReport{
		Metadata: Metadata{
			Hostname:  hostname,
			Timestamp: isoTime,
			ScanID:    fmt.Sprintf("%s_%s", hostname, timestamp),
		},
		Findings: findings,
	}

	if enableCollections {
		fmt.Println("[+] Starting Live Collections (This may take a moment)...")
		telemetry := CollectionData{
			Processes: collections.ProcCollect(),
			Network:   collections.NetCollect(),
			Services:  collections.SvcCollect(),
			Registry:  collections.RegCollect(),
		}
		finalReport.Telemetry = &telemetry
	}

	reportName := fmt.Sprintf("Report_%s_%s.json", hostname, timestamp)

	reportData, err := json.MarshalIndent(finalReport, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal report: %v", err)
		return
	}

	err = os.WriteFile(reportName, reportData, 0644)
	if err != nil {
		log.Printf("Failed to save local report: %v", err)
		return
	}
	fmt.Printf("[+] Local Report Saved: %s\n", reportName)

	reportPath, _ := filepath.Abs(reportName)

	switch reportMode {
	case "Local":
	case "Cloud":
		if sasUrl == "" {
			fmt.Println("[-] Error: Cloud mode requires -SasUrl")
			return
		}
		err := engines.UploadToBlob(reportPath, sasUrl)
		if err != nil {
			log.Printf("[-] Upload Error: %v\n", err)
		}
	case "SMB":
		if smbPath == "" {
			fmt.Println("[-] Error: SMB mode requires -SMBPath")
			return
		}

		destPath := filepath.Join(smbPath, reportName)
		fmt.Printf("[*] Copying to SMB: %s\n", destPath)

		srcFile, err := os.Open(reportPath)
		if err != nil {
			log.Printf("SMB Source Error: %v", err)
			return
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			log.Printf("SMB Dest Error: %v", err)
			return
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			log.Printf("SMB Copy Failed: %v", err)
		} else {
			fmt.Println("[+] SMB Upload Complete")
		}
	}
}

func main() {

	fmt.Println("\n--- DEBUG: ARGUMENT MAP ---")
	for i, arg := range os.Args {
		fmt.Printf("Arg[%d]: '%s'\n", i, arg)
	}
	fmt.Println("---------------------------")

	if len(os.Args) < 2 {
		fmt.Println("Usage: sweeper.exe <config.json> [Local|Cloud|SMB] [SMBPath] [SasUrl] [Collect:true|false]")
		InvokeSweep("sweeper_config.json", "Local", "", "", false)
		return
	}

	jsonPath := os.Args[1]
	mode := "Local"
	smb := ""
	sas := ""
	enableCollections := false

	if len(os.Args) >= 3 {
		mode = os.Args[2]
	}
	if len(os.Args) >= 4 {
		smb = os.Args[3]
	}
	if len(os.Args) >= 5 {
		sas = os.Args[4]
	}
	if len(os.Args) >= 6 {

		argVal := strings.TrimSpace(strings.ToLower(os.Args[5]))
		if argVal == "true" {
			enableCollections = true
		}
	}

	InvokeSweep(jsonPath, mode, smb, sas, enableCollections)
}
