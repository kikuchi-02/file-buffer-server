package libs

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kikuchi-02/file-buffer-server/libs"
	_ "github.com/lib/pq"
)

func TestDb(t *testing.T) {
	libs.LoadDBSettings("../user.yaml")
	db := libs.Connect()
	err := db.Ping()
	if err != nil {
		t.Error(err)
	}
}

func createTracker() libs.Tracker {
	uuid, _ := uuid.NewRandom()
	created := time.Now()
	tracker := libs.Tracker{Uuid: uuid, Created: created}
	return tracker
}

func countRows(db *sql.DB, table string) (int, error) {
	cnt := 1
	var err error
	if table == "tracker" {
		err = db.QueryRow("Select count(*) from okra_core_trackinguser;").Scan(&cnt)
	} else if table == "eventlog" {
		err = db.QueryRow("Select count(*) from okra_core_eventlog;").Scan(&cnt)
	}
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

func TestBulkCreateTracker(t *testing.T) {
	db := libs.Connect()
	defer db.Close()
	before, err := countRows(db, "tracker")
	if err != nil {
		t.Fatal(err)
	}

	trackers := []libs.Tracker{createTracker(), createTracker(), createTracker()}
	libs.BulkCreateTracker(db, &trackers)

	after, err := countRows(db, "tracker")
	if err != nil {
		t.Fatal(err)
	}
	if after-before != 3 {
		t.Errorf("expected 3, but %d, before: %d, after: %d\n", after-before, before, after)
	}

	// duplicate ignored?
	trackers = append([]libs.Tracker{createTracker()}, trackers...)
	libs.BulkCreateTracker(db, &trackers)
	if count, _ := countRows(db, "tracker"); count != after+1 {
		t.Errorf("expected to be same, after: %d, %d\n", count, after)
	}
}

func createEventlog(tracker *libs.Tracker) libs.Eventlog {
	method := "internal"
	eventlog := libs.Eventlog{
		Created:       float32(time.Now().Unix()),
		Tracker:       tracker.Uuid,
		RequestMethod: &method,
	}
	return eventlog
}

func TestBulkCreateEventlog(t *testing.T) {
	db := libs.Connect()
	defer db.Close()
	before, err := countRows(db, "eventlog")
	if err != nil {
		t.Fatal(err)
	}
	trackers := []libs.Tracker{createTracker()}
	libs.BulkCreateTracker(db, &trackers)

	eventlogs := []libs.Eventlog{createEventlog(&trackers[0]),
		createEventlog(&trackers[0]),
		createEventlog(&trackers[0])}

	libs.BulkCreateEventlog(db, &eventlogs)

	after, err := countRows(db, "eventlog")
	if err != nil {
		t.Fatal(err)
	}
	if after-before != 3 {
		t.Errorf("expected 3, but %d, before: %d, after: %d\n", after-before, before, after)
	}
}
