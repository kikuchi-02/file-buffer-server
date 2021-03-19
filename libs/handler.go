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
	UserAgent string      `json:"user_agent"`
	Referrer  string      `json:"referrer"`
	Tracker   uuid.UUID   `json:"tracker"`
	Logs      *[]Eventlog `json:"logs"`
}

type ParsedLogs struct {
	Tracker   *Tracker
	Eventlogs *[]Eventlog
}

func validate(eventlog *Eventlog) bool {
	if eventlog.Created == 0 {
		log.Println("created is 0")
		return false
	}
	if eventlog.Time == 0 {
		log.Println("time is 0")
		return false
	}
	if eventlog.TotalTime == 0 {
		log.Println("total time is 0")
		return false
	}
	return true
}

func Parse(r *http.Request) (*ParsedLogs, error) {
	decoder := json.NewDecoder(r.Body)
	var b RequestBody
	err := decoder.Decode(&b)
	if err != nil {
		return nil, err
	}

	if b.Logs == nil || len(*b.Logs) == 0 {
		return nil, nil
	}

	userAgent := b.UserAgent
	if userAgent == "" {
		return nil, fmt.Errorf("user agent is empty\n")
	}
	if len(userAgent) > 255 {
		userAgent = userAgent[:255]
	}

	var referrer *string
	_referrer := b.Referrer
	if _referrer != "" {
		if len(_referrer) > 255 {
			_referrer = _referrer[:255]
		}
		referrer = &_referrer
	}

	// TODO check
	var country *string
	_country := r.Header.Get("HTTP_CLOUDFRONT_VIEWER_COUNTRY")
	if _country != "" {
		country = &_country
	}
	// request user is authenticated
	// user_id := 0

	tracker_id := b.Tracker
	if tracker_id == uuid.Nil {
		// version 4
		tracker_id, err = uuid.NewRandom()
		if err != nil {
			return nil, err
		}
	}
	trackerTime := time.Now()

	eventlogs := make([]Eventlog, 0, len(*b.Logs))
	for _, eventlog := range *b.Logs {
		if !validate(&eventlog) {
			continue
		}
		eventlog.Tracker = tracker_id
		eventlog.UserAgent = userAgent
		eventlog.Referrer = referrer
		eventlog.Country = country
		eventlogs = append(eventlogs, eventlog)
	}
	if len(eventlogs) == 0 {
		log.Println("No valid log")
		return nil, nil
	}

	parsedLogs := ParsedLogs{Tracker: &Tracker{Uuid: tracker_id, Created: trackerTime}, Eventlogs: &eventlogs}
	return &parsedLogs, nil
}

type ResponseBody struct {
	Tracker uuid.UUID `json:"tracker"`
}

func EventlogHander(source chan *ParsedLogs) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("Bad request method(%s)\n", r.Method)
			w.WriteHeader(404)
			return
		}

		parsed, err := Parse(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		if parsed == nil {
			log.Println("invalid request")
			w.WriteHeader(400)
			return
		}

		response, err := json.Marshal(ResponseBody{Tracker: parsed.Tracker.Uuid})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(response)

		source <- parsed

	}
}
