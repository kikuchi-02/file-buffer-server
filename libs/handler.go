package libs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Eventlog struct {
	RequestMethod string `json:"request_method"`
	// UserAgent     string
	// Referrer      string
	// Country       string
	Place string `json:"place"`
	// should be parsed to time.Time
	Created    float32 `json:"created"`
	Count      int32   `json:"count"`
	Time       float32 `json:"time"`
	TimeStayed float32 `json:"time_stayed"`
	TotalTime  float32 `json:"total_time"`
	// uuid
	Tracker string `json:"tracker_id"`
	User    int32  `json:"user_id"`

	// do not use underscore
	UURL           string  `json:"_user"`
	UURLParams     string  `json:"_user_params"`
	UURLFragment   string  `json:"_url_fragment"`
	UURLParamsHash string  `json:"_url_params_hash"`
	Ucategory      int32   `json:"_category_id"`
	Upost          int32   `json:"_post_id"`
	UmainCategory  int32   `json:"_main_category_id"`
	UcategoryIds   []int32 `json:"_category_ids"`

	Locale        int32             `json:"locale_id"`
	URL           string            `json:"url"`
	URLParams     map[string]string `json:"url_params"`
	URLFragment   string            `json:"url_fragment"`
	URLParamsHash string            `json:"url_params_hash"`
	Categroy      int32             `json:"category_id"`
	Post          int32             `json:"post_id"`
	MainCategory  int32             `json:"main_category_id"`
	CategoryIds   []int32           `json:"category_ids"`
}

type RequestBody struct {
	// UserAgent string     `json:"user_agent"`
	// Referrer  string     `json:"referrer"`
	Logs []Eventlog `json:"logs"`
}

func mapToJsonOrString(obj map[string]string) string {
	bytes, err := json.Marshal(obj)
	if err != nil || bytes == nil {
		return ""
	}
	return string(bytes)
}

func Parse(r *http.Request) (string, error) {
	decoder := json.NewDecoder(r.Body)
	var b RequestBody
	err := decoder.Decode(&b)
	if err != nil {
		return "", err
	}
	if len(b.Logs) == 0 {
		return "", nil
	}

	userAgent := r.UserAgent()
	if len(userAgent) > 255 {
		userAgent = userAgent[:255]
	}
	referrer := r.Referer()
	if len(referrer) > 255 {
		referrer = referrer[:255]
	}
	// TODO
	country := r.Header["HTTP_CLOUDFRONT_VIEWER_COUNTRY"]
	// request user is authenticated
	// tracker is valid? pid?
	tracker_id := b.Logs[0].Tracker
	log.Println(tracker_id)

	formatted := make([]byte, 0, 100)
	for _, log := range b.Logs {
		created := time.Unix(int64(log.Created), 0)
		urlParams := mapToJsonOrString(log.URLParams)
		// TODO all
		str := fmt.Sprintf("%s,%s,%s,%s,%s,%v\n", log.RequestMethod, referrer, created, country, urlParams, log.UcategoryIds)
		formatted = append(formatted, str...)
	}
	log.Println(string(formatted))
	return string(formatted), nil
}

func EventlogHander(source chan string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("Bad request method(%s)\n", r.Method)
			io.WriteString(w, "ok!\n")
			return
		}

		parsed, err := Parse(r)
		if err != nil {
			log.Println(err)
			io.WriteString(w, "ok!\n")
			return
		}
		source <- parsed

		io.WriteString(w, "ok!\n")
	}
}
