package engines

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func UploadToBlob(filePath string, sasUrl string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open report file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %v", err)
	}

	fileName := filepath.Base(filePath)

	parts := strings.SplitN(sasUrl, "?", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid SAS URL format (missing query params)")
	}

	finalURI := fmt.Sprintf("%s/%s?%s", parts[0], fileName, parts[1])

	req, err := http.NewRequest(http.MethodPut, finalURI, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("x-ms-blob-type", "BlockBlob")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload network error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status: %s", resp.Status)
	}

	fmt.Printf("[+] Success: Uploaded %s to Azure Blob\n", fileName)
	return nil
}
