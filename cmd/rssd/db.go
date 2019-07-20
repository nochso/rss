package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/apex/log"
	_ "github.com/mattn/go-sqlite3"
)

var migrations = []Migrator{
	MigrateString(`
CREATE TABLE feed (
    id          INTEGER  PRIMARY KEY
                         NOT NULL,
    title       VARCHAR  NOT NULL,
    link        VARCHAR,
    feed_link   VARCHAR  NOT NULL
                         UNIQUE,
    description VARCHAR,
    language    VARCHAR,
    last_update DATETIME
);

CREATE INDEX idx_feed__last_update ON feed (
    last_update
);

CREATE TABLE feed_item (
    id          INTEGER  PRIMARY KEY
                         NOT NULL,
    feed_id     INTEGER  REFERENCES feed (id) ON DELETE CASCADE
                         NOT NULL,
    title       VARCHAR  NOT NULL,
    link        VARCHAR  NOT NULL,
    published   DATETIME NOT NULL,
    last_update DATETIME NOT NULL
);

CREATE TABLE user_feed_item_bookmark (
    id           INTEGER PRIMARY KEY
                         NOT NULL,
    user_id      INTEGER REFERENCES user (id) ON DELETE CASCADE,
    feed_item_id INTEGER REFERENCES feed_item (id) ON DELETE CASCADE
                         NOT NULL,
    UNIQUE (
        feed_item_id,
        user_id
    )
);

CREATE TABLE user_feed_item_read (
    id           INTEGER PRIMARY KEY
                         NOT NULL,
    user_id      INTEGER REFERENCES user (id) ON DELETE CASCADE,
    feed_item_id INTEGER REFERENCES feed_item (id) ON DELETE CASCADE
                         NOT NULL,
    UNIQUE (
        feed_item_id,
        user_id
    )
);

CREATE TABLE subscription (
	id              PRIMARY KEY
					NOT NULL,
	-- Delete subscriptions when deleting user
	user_id INTEGER REFERENCES user (id) ON DELETE CASCADE
					NOT NULL,
	-- Forbid deleting feeds that have a subscription
	feed_id INTEGER REFERENCES feed (id) ON DELETE RESTRICT
					NOT NULL,
	UNIQUE (
		user_id,
		feed_id
	)
);

CREATE TABLE subscription_tag (
	id              INTEGER PRIMARY KEY
							NOT NULL,
	-- Delete subscription tags when deleting subscription
	subscription_id INTEGER REFERENCES subscription (id) ON DELETE CASCADE
							NOT NULL,
	-- Forbid deleting tags used by a subscription
	tag_id          INTEGER REFERENCES tag (id) ON DELETE RESTRICT
							NOT NULL,
	UNIQUE (
		subscription_id,
		tag_id
	)
);

CREATE TABLE tag (
	id   INTEGER PRIMARY KEY
					NOT NULL,
	name VARCHAR NOT NULL
					UNIQUE
);

CREATE TABLE user (
	id INTEGER NOT NULL
				PRIMARY KEY,
	email VARCHAR NOT NULL
					UNIQUE
);`),
}

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

// Migrator executes a single migration step
type Migrator interface {
	Migrate(*sql.Tx) error
}

// MigrateString is a migration step consisting of one SQL query string
type MigrateString string

// Migrate implements Migrator
func (m MigrateString) Migrate(tx *sql.Tx) error {
	_, err := tx.Exec(string(m))
	return err
}

// MigrateFunc wraps a function that implements Migrator
type MigrateFunc func(*sql.Tx) error

// Migrate implements Migrator
func (m MigrateFunc) Migrate(tx *sql.Tx) error {
	return m(tx)
}

// migrateDB ensures the schema has all migrations applied
func migrateDB(db *sql.DB) error {
	var version int
	err := db.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return err
	}
	if version == len(migrations) {
		return nil
	}
	if version > len(migrations) {
		return fmt.Errorf(
			"db schema version is %d but application only supports version %d",
			version,
			len(migrations),
		)
	}
	log.WithField("current", version).
		WithField("target", len(migrations)).
		Debug("start db migration step")
	for version < len(migrations) {
		mig := migrations[version]
		version++
		start := time.Now()
		err = migrate(db, version, mig)
		if err != nil {
			return fmt.Errorf("migrating db schema to version %d: %v", version, err)
		}
		log.WithField("version", version).
			WithField("target", len(migrations)).
			WithField("duration", time.Since(start)).
			Debug("db migration step complete")
	}
	return nil
}

// migrate db within a single transaction to version using m
func migrate(db *sql.DB, version int, m Migrator) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = m.Migrate(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	// PRAGMA doesn't support ? placeholders
	_, err = db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	return err
}
