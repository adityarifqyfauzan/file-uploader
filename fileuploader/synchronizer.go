package fileuploader

import "sync"

func SynchronizeChunk(chunks []ChunkMeta, metadata map[string]ChunkMeta, uploader Uploader, wg *sync.WaitGroup, mu *sync.Mutex) error {
	// create channel to communicate between goroutines
	chunkChan := make(chan ChunkMeta, len(chunks))
	errChan := make(chan error, len(chunks))

	// iterate over the chunks slice and send each chunk to the chunk channel
	for _, chunk := range chunks {
		wg.Add(1)
		chunkChan <- chunk
	}

	close(chunkChan) // Close the chunk channel to signal that all chunks have been sent

	// start multiple goroutine to process chunks in paralel
	totalWorker := 4
	for i := 0; i < totalWorker; i++ {
		go func() {
			for chunk := range chunkChan {
				defer wg.Done()

				newHash := chunk.MD5Hash

				// check if the chunk exists in the metadata map
				mu.Lock()
				oldChunk, exists := metadata[chunk.FileName]
				mu.Unlock()

				// if the chunk doesn't exists in the metadata map or its hash changed
				if !exists || oldChunk.MD5Hash != newHash {

					// upload the chunk using uploader interface
					err := uploader.UploadChunk(chunk)
					if err != nil {
						errChan <- err
						return
					}

					// update metadata map with the new chunk
					mu.Lock()
					metadata[chunk.FileName] = chunk
					mu.Unlock()

				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Check for errors from the error channel
	for err := range errChan {
		if err != nil {
			return err // Return the first error encountered
		}
	}

	return nil
}
