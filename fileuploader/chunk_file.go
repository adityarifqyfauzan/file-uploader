package fileuploader

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
)

// ChunkFile splits a file into smaller chunks and returns metadata for each chunk.
// It reads the file sequentially and chunks it based on the specified chunk size.
func (c *DefaultFileChunker) ChunkFile(filePath string) ([]ChunkMeta, error) {
	var chunks []ChunkMeta // store metadata for each chunk

	// open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// create a buffer to hold the chunk data
	buffer := make([]byte, c.chunkSize)
	index := 0 // initialize chunk index

	// loop until EOF is reached
	for {
		// read chunkSize bytes from the file into the buffer
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if bytesRead == 0 {
			break // If bytesRead is 0, it means EOF is reached
		}

		// generate a unique hash for the chunk data
		hash := md5.Sum(buffer[:bytesRead])
		hashString := hex.EncodeToString(hash[:])

		// construct the chunk file name
		chunkFileName := fmt.Sprintf("%s.chunk.%d", filePath, index)

		// create a new chunk file and write the buffer data to it
		chunkFile, err := os.Create(chunkFileName)
		if err != nil {
			return nil, err
		}
		_, err = chunkFile.Write(buffer[:bytesRead])
		if err != nil {
			return nil, err
		}

		// append metadata for the chunk to the chunks slice
		chunks = append(chunks, ChunkMeta{FileName: chunkFileName, MD5Hash: hashString, Index: index})

		chunkFile.Close()
		index++
	}

	return chunks, nil
}

// ChunklargeFile splits a large file into smaller chunks in parallel and returns metadata for each chunk.
// It divides the file into chunks and processes them concurrently using multiple goroutines.
func (c *DefaultFileChunker) ChunkLargeFile(filePath string) ([]ChunkMeta, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var chunks []ChunkMeta

	// open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// get file info to determine the number of chunks
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	numChunks := int(fileInfo.Size() / int64(c.chunkSize))
	if fileInfo.Size()%int64(c.chunkSize) != 0 {
		numChunks++
	}

	// create channels to communicate between goroutines
	chunkChan := make(chan ChunkMeta, numChunks)
	errChan := make(chan error, numChunks)
	indexChan := make(chan int, numChunks)

	// populate the index channel with chunk indices
	for i := 0; i < numChunks; i++ {
		indexChan <- i
	}
	close(indexChan)

	// start multiple goroutines to process chunks in parallel
	totalWorker := 4
	for i := 0; i < totalWorker; i++ {
		wg.Add(1)
		go func() {
			for index := range indexChan {

				// calculate the offset for the current chunk
				offset := index * c.chunkSize
				buffer := make([]byte, c.chunkSize)

				// seek to the appropiate position in the file
				file.Seek(int64(offset), 0)

				// read chunkSize bytes from the file into the buffer
				bytesRead, err := file.Read(buffer)
				if err != nil && err != io.EOF {
					errChan <- err
					return
				}

				// if bytesRead is 0, it means EOF is reached
				if bytesRead > 0 {
					// generate a unique hash for the chunk data
					hash := md5.Sum(buffer[:bytesRead])
					hashString := hex.EncodeToString(hash[:])

					// construct the chunk file name
					chunkFileName := fmt.Sprintf("%s.chunk.%d", filePath, index)

					// create a new chunk file and write the buffer data to it
					chunkFile, err := os.Create(chunkFileName)
					if err != nil {
						errChan <- err
						return
					}
					_, err = chunkFile.Write(buffer[:bytesRead])
					if err != nil {
						errChan <- err
						return
					}

					// append metadata for the chunk to the chunks slice
					chunk := ChunkMeta{
						FileName: chunkFileName,
						MD5Hash:  hashString,
						Index:    index,
					}
					mu.Lock()
					chunks = append(chunks, chunk)
					mu.Unlock()

					chunkFile.Close()

					// send the processed chunk to the chunk channel
					chunkChan <- chunk
				}

			}
		}()
	}

	go func() {
		wg.Wait()
		close(chunkChan)
		close(errChan)
	}()

	// Check for errors from goroutines
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return chunks, nil
}
