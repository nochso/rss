package main

import (
	"database/sql"

	"github.com/apex/log"
	_ "github.com/mattn/go-sqlite3"
)

func openDB(fpath string) (*sql.DB, error) {
	// use write-ahead log and wait 10s when locked
	dsn := "file:" + fpath + "?_journal=WAL&_synchronous=NORMAL&_busy_timeout=10000"
	log.WithField("dsn", dsn).
		WithField("file", fpath).
		Debug("opening db")
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`PRAGMA foreign_keys = on;`)
	if err != nil {
		return nil, err
	}
	err = migrateDB(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func closeDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.WithError(err).Error("closing db")
		return
	}
	log.Info("db closed")
}
