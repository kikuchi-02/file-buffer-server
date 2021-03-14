package libs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sync"
	"time"
)

const (
	Dirname  = "tmp"
	Megabyte = 1 << 20
	// MaxFileSize = Megabyte * 10
	MaxFileSize      = 1024
	MaxPassedMinutes = 30
	Concurrency      = 5
)

/* Check is a mutex is locked
https://blog.trailofbits.com/2020/06/09/how-to-check-if-a-mutex-is-locked-in-go/
*/
func MutexLocked(m *sync.Mutex) bool {
	mutexLocked := int64(1)
	state := reflect.ValueOf(m).Elem().FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

func RunCopy(filePath string) error {
	// time.Sleep(time.Second * 100)
	commandArgs := []string{
		"3",
	}
	out, err := exec.Command("sleep", commandArgs...).Output()
	if err != nil {
		return err
	}
	log.Printf("Output: %s", string(out))
	return nil
}

func CleanUpFile(filePath string) *os.File {
	if err := os.Remove(filePath); err != nil {
		log.Println(err)
		return nil
	}
	// new file
	file, err := os.Create(filePath)
	if err != nil {
		log.Println(err)
		return nil
	}
	return file
}

func worker(source chan string, id int, mutex *sync.Mutex) {
	startTime := time.Now()
	filePath := filepath.Join(Dirname, fmt.Sprintf("buffer-%d.log", id))
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Println(err)
		panic(err)
	}

	for req := range source {
		// write to file
		_, err := file.WriteString(req + "\n")
		if err != nil {
			log.Println(err)
			continue
		}
		// get file info for size comparing
		fileInfo, err := file.Stat()
		if err != nil {
			log.Println(err)
			continue
		}
		// passed time from start
		passedTime := time.Now().Sub(startTime).Minutes()

		if fileInfo.Size() > MaxFileSize || passedTime > MaxPassedMinutes {
			if MutexLocked(mutex) {
				continue
			}
			mutex.Lock()

			file.Close()

			log.Println("filePath", filePath, "filesize", fileInfo.Size()>>20, "Mb", "passed time", passedTime)
			err := RunCopy(filePath)
			if err == nil {
				newFile := CleanUpFile(filePath)
				if newFile != nil {
					file = newFile
				}
			} else {
				log.Printf("Copy command raised Error %v", err)
			}

			mutex.Unlock()
			startTime = time.Now()
		}
	}
}

func BufferSetup() chan string {
	// buffer dir
	if _, err := os.Stat(Dirname); os.IsNotExist(err) {
		log.Println("Create buffer directory")
		os.Mkdir(Dirname, 0777)
	}
	source := make(chan string)
	mutex := sync.Mutex{}
	// thread pool
	for i := 0; i < Concurrency; i++ {
		go worker(source, i, &mutex)
	}
	return source
}
