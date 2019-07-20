# rss
RSS feed aggregator/reader

## Installation

### From source

    go get github.com/nochso/rss
    cd %GOPATH%/src/github.com/nochso/rss
    go build ./cmd/rssd
    rssd

## Usage

    $ rssd -h
    Usage of rssd:
      -db string
            sqlite3 db file (default "rss.sqlite3")
      -grace duration
            HTTP shutdown grace period for existing connections (default 10s)
      -http string
            HTTP listening address (default ":8080")
