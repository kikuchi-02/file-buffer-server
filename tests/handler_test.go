package libs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kikuchi-02/file-buffer-server/libs"
)

func assert(t *testing.T, expect interface{}, data interface{}) {
	if expect != data {
		t.Errorf("expected: %v, but %v\n", expect, data)
	}
}

func assertNil(t *testing.T, data interface{}) {
	if data == nil {
		t.Errorf("expected to be nil, but %v\n", data)
	}
}

func assertSlice(t *testing.T, expect []int32, data []int32) {
	if len(expect) != len(data) {
		t.Errorf("expected: %d, but %d", len(expect), len(data))
	}
	for i, v := range expect {
		if v != data[i] {
			t.Errorf("expected: %d, but %d", v, data[i])
		}
	}
}

func assertStruct(t *testing.T, expect interface{}, data interface{}) {
	// 他の方法でうまく行かなかった。
	if fmt.Sprint(expect) != fmt.Sprint(data) {
		t.Errorf("expected: %v, but %v\n", expect, data)
	}
}

func TestParse(t *testing.T) {
	uuid, _ := uuid.NewRandom()
	body := bytes.NewBuffer([]byte(fmt.Sprintf(`
	{
		"user_agent": "test-agent",
		"referrer": "test-referrer",
		"tracker": "%s",
		"logs": [
					{
	 					"request_method": "method",
	 					"place": "place",
						"created": 1615729350740,
						 "count": 3,
						 "time": 10.5,
						 "time_stayed": 1.5,
						 "total_time": 1.4,
						 "_url": "/url",
						 "_url_fragment": "fragment",
						 "_url_params": { "page": "params" },
						 "_url_params_hash": "hash",
						 "_category": 1,
						 "_post": 1,
						 "_main_category": 1,
						 "_category_ids": [1, 2],
						 "url": "/url",
						 "url_fragment": "fragment",
						 "url_params": { "page": "params" },
						 "url_params_hash": "hash",
						 "locale": 1,
						 "category": 1,
						 "post": 1,
						 "main_category": 1,
						 "category_ids": [1, 2]
					}
		]
	}
	`, uuid)))

	request := httptest.NewRequest(http.MethodPost, "http://localhost:8000/eventlog", body)
	request.Header.Add("HTTP_CLOUDFRONT_VIEWER_COUNTRY", "test-country")
	response, err := libs.Parse(request)
	if err != nil {
		t.Error(err)
	}
	if len(response.Logs) != 1 {
		t.Errorf("expected %d, but %d", 1, len(response.Logs))
	}
	log := (response.Logs)[0]
	urlParams := map[string]string{
		"page": "params",
	}

	assert(t, "test-agent", response.UserAgent)
	assert(t, "test-referrer", *response.Referrer)
	assert(t, "test-country", *response.Country)

	assert(t, "method", *log.RequestMethod)
	assertNil(t, log.UserAgent)
	assertNil(t, log.Referrer)
	assertNil(t, log.Country)
	assert(t, "place", *log.Place)
	assert(t, float32(1615729350740), log.Created)
	assert(t, 3, *log.Count)
	assert(t, float32(10.5), log.Time)
	assert(t, float32(1.5), *log.TimeStayed)
	assert(t, float32(1.4), log.TotalTime)

	assert(t, "/url", *log.UURL)
	assert(t, "fragment", *log.UURLFragment)
	assertStruct(t, urlParams, log.UURLParams)
	assert(t, "hash", *log.UURLParamsHash)
	assert(t, 1, *log.Ucategory)
	assert(t, 1, *log.Upost)
	assert(t, 1, *log.UmainCategory)
	assertSlice(t, []int32{1, 2}, *log.UcategoryIds)

	assert(t, 1, *log.Locale)
	assert(t, "/url", *log.URL)
	assert(t, "fragment", *log.URLFragment)
	assertStruct(t, urlParams, log.URLParams)
	assert(t, "hash", *log.URLParamsHash)
	assert(t, 1, *log.Category)
	assert(t, 1, *log.Post)
	assert(t, 1, *log.MainCategory)
	assertSlice(t, []int32{1, 2}, *log.CategoryIds)

}

func parseResponse(r *httptest.ResponseRecorder) (*libs.RequestBody, error) {
	var res libs.RequestBody
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func TestEventlogHandler(t *testing.T) {
	source := libs.BufferSetup()

	body := bytes.NewBuffer([]byte(`
	{
	}
	`))
	// get
	request := httptest.NewRequest(http.MethodGet, "http://localhost:8000/eventlog", body)
	response := httptest.NewRecorder()
	libs.EventlogHander(source)(response, request)
	if response.Code != 404 {
		t.Errorf("expected 404, but %d\n", response.Code)
	}

	// invalid format
	body = bytes.NewBuffer([]byte(`
	{
		"
	}
	`))
	request = httptest.NewRequest(http.MethodPost, "http://localhost:8000/eventlog", body)
	response = httptest.NewRecorder()
	libs.EventlogHander(source)(response, request)
	if response.Code != 400 {
		t.Errorf("expected 400, but %d\n", response.Code)
	}

	// inadequate params
	body = bytes.NewBuffer([]byte(`
	{
		"user-agent": "test"
	}
	`))
	request = httptest.NewRequest(http.MethodPost, "http://localhost:8000/eventlog", body)
	response = httptest.NewRecorder()
	libs.EventlogHander(source)(response, request)
	if response.Code != 400 {
		t.Errorf("expected 400, but %d\n", response.Code)
	}

	// no tracker
	body = bytes.NewBuffer([]byte(`
	{
		"user_agent": "test-agent",
		"referrer": "test-referrer",
		"logs": [
					{
	 					"request_method": "method",
	 					"place": "place",
						"created": 1615729350740,
						 "count": 3,
						 "time": 10.5,
						 "time_stayed": 1.5,
						 "total_time": 1.4,
						 "_url": "/url",
						 "_url_fragment": "fragment",
						 "_url_params": { "page": "params" },
						 "_url_params_hash": "hash",
						 "_category": 1,
						 "_post": 1,
						 "_main_category": 1,
						 "_category_ids": [1, 2],
						 "url": "/url",
						 "url_fragment": "fragment",
						 "url_params": { "page": "params" },
						 "url_params_hash": "hash",
						 "locale": 1,
						 "category": 1,
						 "post": 1,
						 "main_category": 1,
						 "category_ids": [1, 2]
					}
		]
	}
	`))
	request = httptest.NewRequest(http.MethodPost, "http://localhost:8000/eventlog", body)
	response = httptest.NewRecorder()
	libs.EventlogHander(source)(response, request)

	if response.Code != 201 {
		t.Errorf("expected 201, but %d", response.Code)
	}
	res, err := parseResponse(response)
	if err != nil {
		t.Error(err)
	}
	if res.TrackerId == uuid.Nil {
		t.Error("tracker id should not be nil")
	}

	// with tracker
	tracker, _ := uuid.NewRandom()
	body = bytes.NewBuffer([]byte(fmt.Sprintf(`
	{
		"user_agent": "test-agent",
		"referrer": "test-referrer",
		"tracker": "%s",
		"logs": [
					{
	 					"request_method": "method",
	 					"place": "place",
						"created": 1615729350740,
						 "count": 3,
						 "time": 10.5,
						 "time_stayed": 1.5,
						 "total_time": 1.4,
						 "_url": "/url",
						 "_url_fragment": "fragment",
						 "_url_params": { "page": "params" },
						 "_url_params_hash": "hash",
						 "_category": 1,
						 "_post": 1,
						 "_main_category": 1,
						 "_category_ids": [1, 2],
						 "url": "/url",
						 "url_fragment": "fragment",
						 "url_params": { "page": "params" },
						 "url_params_hash": "hash",
						 "locale": 1,
						 "category": 1,
						 "post": 1,
						 "main_category": 1,
						 "category_ids": [1, 2]
					}
		]
	}
	`, tracker)))
	request = httptest.NewRequest(http.MethodPost, "http://localhost:8000/eventlog", body)
	response = httptest.NewRecorder()
	libs.EventlogHander(source)(response, request)

	if response.Code != 201 {
		t.Errorf("expected 201, but %d", response.Code)
	}
	res, err = parseResponse(response)
	if err != nil {
		t.Error(err)
	}
	if res.TrackerId != tracker {
		t.Error("tracker id should be same")
	}
}
