package libs

import (
	"log"
	"reflect"
	"sync"
	"time"
)

/* Check if a mutex is locked
https://blog.trailofbits.com/2020/06/09/how-to-check-if-a-mutex-is-locked-in-go/
*/
func MutexLocked(m *sync.Mutex) bool {
	mutexLocked := int64(1)
	state := reflect.ValueOf(m).Elem().FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

func includes(trackers []Tracker, tracker Tracker) bool {
	for _, t := range trackers {
		if t.Uuid == tracker.Uuid {
			return true
		}
	}
	return false
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

func worker(source chan *RequestBody, id int, mutex *sync.Mutex) {
	startTime := time.Now()
	eventlogs := make([]Eventlog, 0, MaxArrayLength*1.5)
	trackers := make([]Tracker, 0, MaxArrayLength*1.5)

	for req := range source {

		for _, log := range req.Logs {
			if !validate(&log) {
				continue
			}
			log.Tracker = req.TrackerId
			log.UserAgent = req.UserAgent
			log.Referrer = req.Referrer
			log.Country = req.Country
			eventlogs = append(eventlogs, log)
		}
		tracker := Tracker{req.TrackerId, req.TrackerCreated}
		if !includes(trackers, tracker) {
			trackers = append(trackers, tracker)
		}

		// passed time from start
		passedTime := time.Now().Sub(startTime).Minutes()

		if MutexLocked(mutex) {
			continue
		}
		if len(eventlogs) > MaxArrayLength || passedTime > MaxPassedMinutes {
			// 上のラインを計算している間にlockされることがある。
			if MutexLocked(mutex) {
				continue
			}
			mutex.Lock()

			log.Println("buffer length", len(eventlogs), "passed time", passedTime)

			db := Connect()

			err := BulkCreateTracker(db, &trackers)
			if err != nil {
				log.Println(err)
			} else {
				err = BulkCreateEventlog(db, &eventlogs)
				if err != nil {
					log.Panicln(err)
				}
			}

			if passedTime > MaxPassedMinutes {
				trackers = make([]Tracker, 0, MaxArrayLength*1.5)
				eventlogs = make([]Eventlog, 0, MaxArrayLength*1.5)
			} else {
				// keep capacity
				trackers = trackers[:0]
				eventlogs = eventlogs[:0]

			}

			if err := db.Close(); err != nil {
				log.Println(err)
			}

			mutex.Unlock()
			startTime = time.Now()
		}
	}
}

func BufferSetup() chan *RequestBody {
	source := make(chan *RequestBody)
	mutex := sync.Mutex{}
	// thread pool
	for i := 0; i < Concurrency; i++ {
		go worker(source, i, &mutex)
	}
	return source
}
