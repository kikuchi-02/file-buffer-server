package libs

import (
	"log"
	"reflect"
	"sync"
	"time"
)

const (
	// RESが30mbいかない程度の設定。
	MaxArrayLength   = 1e3
	MaxPassedMinutes = 30
	// more than 2
	Concurrency = 5
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

func worker(source chan *ParsedLogs, id int, mutex *sync.Mutex) {
	startTime := time.Now()
	eventlogs := make([]Eventlog, 0, MaxArrayLength*1.5)
	trackers := make([]Tracker, 0, MaxArrayLength*1.5)

	for req := range source {

		eventlogs = append(eventlogs, *req.Eventlogs...)
		if !includes(trackers, *req.Tracker) {
			trackers = append(trackers, *req.Tracker)
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

			// capacityは起動時に肥大することがあるのでリセットする。
			BulkCreateTracker(db, &trackers)
			trackers = make([]Tracker, 0, MaxArrayLength*1.5)
			BulkCreateEventlog(db, &eventlogs)
			eventlogs = make([]Eventlog, 0, MaxArrayLength*1.5)

			if err := db.Close(); err != nil {
				log.Println(err)
			}

			mutex.Unlock()
			startTime = time.Now()
		}
	}
}

func BufferSetup() chan *ParsedLogs {
	source := make(chan *ParsedLogs)
	mutex := sync.Mutex{}
	// thread pool
	for i := 0; i < Concurrency; i++ {
		go worker(source, i, &mutex)
	}
	return source
}
