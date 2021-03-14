package libs

import (
	"io/ioutil"
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

func TestRunCopy(t *testing.T) {
	filePath := "template.log"
	content := "test"
	err := createFile(filePath, content)
	if err != nil {
		t.Fatalf("Cannot create file %#v", err)
	}
	err = libs.RunCopy(filePath)
	if err != nil {
		t.Errorf("Error raised %v\n", err)
	}
	os.Remove(filePath)
}

func TestCleanUpFile(t *testing.T) {
	filePath := "template.log"
	content := "test"
	err := createFile(filePath, content)
	if err != nil {
		t.Fatalf("Cannot create file %#v", err)
	}

	file := libs.CleanUpFile(filePath)
	if file == nil {
		t.Error("File should be returned")
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal("Cound not read file")
	}
	if len(b) > 0 {
		t.Error("File should be emtpy")
	}
	os.Remove(filePath)
}
