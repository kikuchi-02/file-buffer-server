package libs

import (
	"os"
	"testing"

	"github.com/kikuchi-02/file-buffer-server/libs"
)

func TestMain(m *testing.M) {
	libs.LoadDBSettings("../user.yaml")
	code := m.Run()
	os.Exit(code)
}
