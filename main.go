package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/adityarifqyfauzan/file-uploader/fileuploader"
	"github.com/joho/godotenv"
)

const (
	defaultChunkSize = 1024 * 1024 // 1MB chunks
	maxRetries       = 3
)

func loadEnv() error {
	return godotenv.Load()
}

func main() {
	err := loadEnv()
	if err != nil {
		log.Println("no .env file found, using default config")
	}

	chunkSize := defaultChunkSize
	if size, ok := os.LookupEnv("CHUNK_SIZE"); ok {
		fmt.Sscanf(size, "%d", &chunkSize)
	}

	serverURL, ok := os.LookupEnv("SERVER_URL")
	if !ok {
		log.Fatal("SERVER_URL env is required!")
	}

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <file_path>")
	}
	filePath := os.Args[1]

	config := fileuploader.Config{ChunkSize: chunkSize, ServerURL: serverURL}

	chunker := &fileuploader.DefaultFileChunker{}
	uploader := &fileuploader.DefaultUploader{}
	metadataManager := &fileuploader.DefaultMetadataManager{}

	chunker.SetChunkSize(config.ChunkSize)
	uploader.SetServerURL(config.ServerURL)

	chunks, err := chunker.ChunkFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	metadata, err := metadataManager.LoadMetadata(fmt.Sprintf("%s.metadata.json", filePath))
	if err != nil {
		log.Println("Could not load metadata, starting fresh")
		metadata = make(map[string]fileuploader.ChunkMeta)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	err = fileuploader.SynchronizeChunk(chunks, metadata, uploader, &wg, &mu)
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	err = metadataManager.SaveMetadata(fmt.Sprintf("%s.metadata.json", filePath), metadata)
	if err != nil {
		log.Fatal(err)
	}

	changeChan := make(chan bool)
	go fileuploader.WatchFile(filePath, changeChan)

	for {
		select {
		case <-changeChan:
			log.Println("file changed, re-chunking and synchronizing...")
			chunks, err := chunker.ChunkFile(filePath)
			if err != nil {
				log.Fatal(err)
			}

			err = fileuploader.SynchronizeChunk(chunks, metadata, uploader, &wg, &mu)
			if err != nil {
				log.Fatal(err)
			}

			wg.Wait()

			err = metadataManager.SaveMetadata(fmt.Sprintf("%s.metadata.json", filePath), metadata)
			if err != nil {
				log.Fatal(err)
			}

		case <-time.After(10 * time.Second):
			log.Println("no changes detected, checking again...")
		}
	}
}
