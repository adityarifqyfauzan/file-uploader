package fileuploader

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

func (u *DefaultUploader) UploadChunk(chunk ChunkMeta) error {
	data, err := os.ReadFile(chunk.FileName)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.serverURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload chunk: %s", resp.Status)
	}

	return nil
}
