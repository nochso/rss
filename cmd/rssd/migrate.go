package main

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
    guid        VARCHAR  NOT NULL,
    title       VARCHAR  NOT NULL,
    link        VARCHAR  NOT NULL,
    published   DATETIME NOT NULL,
    last_update DATETIME NOT NULL,
    UNIQUE (
        guid,
        feed_id
    )
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
    id    INTEGER NOT NULL
                  PRIMARY KEY,
    email VARCHAR NOT NULL
                  UNIQUE
);`),
}
