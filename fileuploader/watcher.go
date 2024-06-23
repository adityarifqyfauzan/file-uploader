package fileuploader

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func WatchFile(filePath string, changeChan chan bool) {
	// create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err) // terminate the program if an error occurs while creating the watcher
	}
	defer watcher.Close()

	// add specified file to the watcher's list of watch files
	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// infinite loop to continously monitor events from the watcher
	for {
		select {
		case event, ok := <-watcher.Events:
			// check if the event channel is closed
			if !ok {
				return // exit the function if channel is closed
			}

			// check if the event corresponds to a write operation on the file
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)
				changeChan <- true
			}

		case err, ok := <-watcher.Errors:
			// check if the errors channel is closed
			if !ok {
				return // exit the function
			}
			log.Println("error:", err)
		}
	}
}
