package libs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Tracker struct {
	Uuid    uuid.UUID
	Created time.Time
}

type Eventlog struct {
	RequestMethod *string `json:"request_method"`
	// not null
	UserAgent string
	Referrer  *string
	Country   *string
	Place     *string `json:"place"`
	// should be parsed to to_timestamp
	// not null
	Created float32 `json:"created"`
	// Created    time.Time `json:"created,timestamp"`
	Count *int `json:"count"`
	// not null
	Time       float32  `json:"time"`
	TimeStayed *float32 `json:"time_stayed"`
	// not null
	TotalTime float32 `json:"total_time"`
	Tracker   uuid.UUID
	User      *int `json:"user"`

	// do not use underscore
	UURL           *string     `json:"_url"`
	UURLParams     interface{} `json:"_url_params"`
	UURLFragment   *string     `json:"_url_fragment"`
	UURLParamsHash *string     `json:"_url_params_hash"`
	Ucategory      *int        `json:"_category"`
	Upost          *int        `json:"_post"`
	UmainCategory  *int        `json:"_main_category"`
	UcategoryIds   *[]int32    `json:"_category_ids"`

	Locale        *int        `json:"locale"`
	URL           *string     `json:"url"`
	URLParams     interface{} `json:"url_params"`
	URLFragment   *string     `json:"url_fragment"`
	URLParamsHash *string     `json:"url_params_hash"`
	Categroy      *int        `json:"category"`
	Post          *int        `json:"post"`
	MainCategory  *int        `json:"main_category"`
	CategoryIds   *[]int32    `json:"category_ids"`
}

func Connect() *sql.DB {
	db, err := sql.Open("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME")))

	if err != nil {
		log.Fatal(err)
	}
	return db
}

func bulkCreate(db *sql.DB, runStmt func(txn *sql.Tx) *sql.Stmt) {
	txn, err := db.Begin()
	if err != nil {
		log.Println(err)
		return
	}

	stmt := runStmt(txn)

	err = stmt.Close()
	if err != nil {
		log.Println(err)
		return
	}
	err = txn.Commit()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Successfully created")
}

func BulkCreateTracker(db *sql.DB, trackers *[]Tracker) {
	bulkCreate(db, func(txn *sql.Tx) *sql.Stmt {
		stmt, err := txn.Prepare(`
		with T (uuid, created) as (values ($1, $2))
		INSERT INTO okra_core_trackinguser( uuid, created )
		SELECT uuid::uuid, created::timestamptz
		FROM T
		WHERE NOT EXISTS (
			SELECT 1
			FROM okra_core_trackinguser
			WHERE T.uuid::uuid = okra_core_trackinguser.uuid
		)
		`)
		if err != nil {
			log.Println(err)
			return stmt
		}
		for _, tracker := range *trackers {
			_, err := stmt.Exec(tracker.Uuid, tracker.Created)
			if err != nil {
				log.Println(err)
				return stmt
			}
		}
		return stmt

	})
}

func toJsonOrString(obj interface{}) *string {
	bytes, err := json.Marshal(obj)
	if err != nil || bytes == nil {
		return nil
	}
	str := string(bytes)
	return &str
}

func unixToTime(unixTime float32) time.Time {
	return time.Unix(int64(unixTime), 0)
}

// 外部キー制約などに引っかかるとすべてrollbackされることに注意。
func BulkCreateEventlog(db *sql.DB, eventlogs *[]Eventlog) {
	bulkCreate(db, func(txn *sql.Tx) *sql.Stmt {
		stmt, err := txn.Prepare(pq.CopyIn(
			"okra_core_eventlog",
			"request_method", "user_agent", "referrer",
			"place", "created", "count",
			"country", "time", "time_stayed",
			"total_time", "_url", "_url_params",
			"_url_fragment", "_url_params_hash", "_category_ids",
			"url", "url_params", "url_fragment",
			"url_params_hash", "category_ids", "_category_id",
			"_main_category_id", "_post_id", "category_id",
			"locale_id", "main_category_id", "post_id",
			"tracker_id", "user_id",
		))
		if err != nil {
			log.Println(err)
			return stmt
		}
		for _, eventlog := range *eventlogs {
			_, err = stmt.Exec(
				eventlog.RequestMethod, eventlog.UserAgent, eventlog.Referrer,
				eventlog.Place, unixToTime(eventlog.Created), eventlog.Count,
				eventlog.Country, eventlog.Time, eventlog.TimeStayed,
				eventlog.TotalTime, eventlog.UURL, toJsonOrString(eventlog.UURLParams),
				eventlog.UURLFragment, eventlog.UURLParamsHash, pq.Array(eventlog.UcategoryIds),
				eventlog.URL, toJsonOrString(eventlog.URLParams), eventlog.URLFragment,
				eventlog.URLParamsHash, pq.Array(eventlog.CategoryIds), eventlog.Ucategory,
				eventlog.UmainCategory, eventlog.Post, eventlog.Categroy,
				eventlog.Locale, eventlog.MainCategory, eventlog.Post,
				eventlog.Tracker, eventlog.User,
			)
			if err != nil {
				log.Println(err)
				return stmt
			}
		}
		_, err = stmt.Exec()
		if err != nil {
			log.Println(err)
			return stmt
		}
		return stmt

	})
}
