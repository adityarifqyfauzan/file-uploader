package fileuploader

import (
	"encoding/json"
	"os"
)

func (m *DefaultMetadataManager) LoadMetadata(filePath string) (map[string]ChunkMeta, error) {
	metadata := make(map[string]ChunkMeta)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}

func (m *DefaultMetadataManager) SaveMetadata(filePath string, metadata map[string]ChunkMeta) error {
	data, err := json.MarshalIndent(metadata, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
