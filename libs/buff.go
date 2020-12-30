package libs

import (
	"fmt"
	"os"
	"time"
)

const (
	Dirname          = "tmp"
	Megabyte         = 1 << 20
	MaxFileSize      = Megabyte * 10
	MaxPassedMinutes = 30
	Concurrency      = 5
)

func worker(source chan string, id int, isProcessing *bool) {
	startTime := time.Now()
	filePath := fmt.Sprintf("%s/buffer-%d.txt", Dirname, id)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// defer file.Close()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	for req := range source {
		// write to file
		_, err := file.WriteString(req + "\n")
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		// get file info for size comparing
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		// passed time from start
		passedTime := time.Now().Sub(startTime).Minutes()

		if !*isProcessing && (fileInfo.Size() > MaxFileSize || passedTime > MaxPassedMinutes) {
			*isProcessing = true
			fmt.Println("filePath", filePath, "filesize", fileInfo.Size()>>20, "Mb", "passed time", passedTime)

			// simething heavy func
			time.Sleep(time.Second * 3)

			*isProcessing = false

			file.Close()
			// new file
			file, err = os.Create(filePath)
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			startTime = time.Now()
		}
	}
}

func BufferSetup() chan string {
	// buffer dir
	if _, err := os.Stat(Dirname); os.IsNotExist(err) {
		fmt.Println(err)
		os.Mkdir(Dirname, 0777)
	}
	source := make(chan string)
	isProcessing := false
	// thread pool
	for i := 0; i < Concurrency; i++ {
		go worker(source, i, &isProcessing)
	}
	return source
}
