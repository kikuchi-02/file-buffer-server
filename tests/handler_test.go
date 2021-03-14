package libs

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kikuchi-02/file-buffer-server/libs"
)

func TestParse(t *testing.T) {
	body := bytes.NewBuffer([]byte(`
	{
		"user_agent": "test-agent",
		"referrer": "test-referrer",
		"logs": [
			{
				"request_method": "internal",
				"url_params": {"page": "4"},
				"_category_ids": [613, 1772, 184],
				"created": 1615729350740
			},
			{
				"request_method": "visit"
			}
		]
	}
	`))
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8000/__api/eventlogs", body)
	response, err := libs.Parse(request)
	if err != nil {
		t.Error(err)
	}
	log.Println(response)
}

func TestEventlogHandler(t *testing.T) {
	body := bytes.NewBuffer([]byte(`
	{
		"RrequestMethod": "test"
	}
	`))
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8000/__api/eventlogs", body)
	response := httptest.NewRecorder()

	source := libs.BufferSetup()
	libs.EventlogHander(source)(response, request)

	log.Println("code", response.Code)
	log.Println("body", response.Body)

	if response.Code != 200 {
		t.Errorf("exptected 200, but %d", response.Code)
	}

	os.RemoveAll(libs.Dirname)
}
