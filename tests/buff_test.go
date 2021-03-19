package libs

import (
	"os"
	"sync"
	"testing"

	"github.com/kikuchi-02/file-buffer-server/libs"
)

func createFile(filePath string, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

func TestMutexLocked(t *testing.T) {
	mutex := sync.Mutex{}
	if libs.MutexLocked(&mutex) {
		t.Error("Mutex should be locked")
	}
	mutex.Lock()
	if !libs.MutexLocked(&mutex) {
		t.Error("Mutex should be locked")
	}
	mutex.Unlock()
	if libs.MutexLocked(&mutex) {
		t.Error("Mutex shoud not be locked")
	}
}
