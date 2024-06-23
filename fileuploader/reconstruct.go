package fileuploader

import (
	"io"
	"os"
	"sort"
)

// ReconstructFile reconstructs a file from its chunks using the provided metadata.
// It reads each chunk file, concatenates their content, and writes it to the output file.
func ReconstructFile(metadata map[string]ChunkMeta, outputFilePath string) error {
	// create or truncate the output file
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// iterate through the metadata to determine the order of the chunks
	var chunks []ChunkMeta
	for _, chunk := range metadata {
		chunks = append(chunks, chunk)
	}

	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Index < chunks[j].Index
	})

	// iterate through the sorted chunks and concatenate their content to reconstruct the file
	for _, chunk := range chunks {
		// open chunk file
		chunkFile, err := os.Open(chunk.FileName)
		if err != nil {
			return err
		}
		defer chunkFile.Close()

		// read the content of the chunk file and write it to the output file
		_, err = io.Copy(outputFile, chunkFile)
		if err != nil {
			return err
		}
	}

	return nil
}
