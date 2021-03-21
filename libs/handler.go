package libs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type RequestBody struct {
	UserAgent string  `json:"user_agent"`
	Referrer  *string `json:"referrer"`
	// from header
	Country        *string
	TrackerId      uuid.UUID `json:"tracker"`
	TrackerCreated time.Time
	Logs           []Eventlog `json:"logs"`
}

type ResponseBody struct {
	Tracker uuid.UUID `json:"tracker"`
}

func Parse(r *http.Request) (*RequestBody, error) {
	decoder := json.NewDecoder(r.Body)
	var b RequestBody
	err := decoder.Decode(&b)
	if err != nil {
		return nil, err
	}

	if len(b.Logs) == 0 {
		return nil, fmt.Errorf("no logs\n")
	}

	if b.UserAgent == "" {
		return nil, fmt.Errorf("user agent is empty\n")
	}
	if len(b.UserAgent) > 255 {
		b.UserAgent = b.UserAgent[:255]
	}

	if b.Referrer != nil {
		if len(*b.Referrer) > 255 {
			*b.Referrer = (*b.Referrer)[:255]
		}
	}

	country := r.Header.Get("HTTP_CLOUDFRONT_VIEWER_COUNTRY")
	if country != "" {
		b.Country = &country
	}
	// request user is authenticated
	// user_id := 0

	if b.TrackerId == uuid.Nil {
		// version 4
		b.TrackerId, err = uuid.NewRandom()
		if err != nil {
			return nil, err
		}
	}
	b.TrackerCreated = time.Now()

	return &b, nil
}

func EventlogHander(source chan *RequestBody) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("Bad request method(%s)\n", r.Method)
			w.WriteHeader(404)
			return
		}

		parsed, err := Parse(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(400)
			return
		}

		response, err := json.Marshal(ResponseBody{Tracker: parsed.TrackerId})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(response)

		source <- parsed

	}
}
