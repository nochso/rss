# rss
RSS feed aggregator/reader

## Installation

### From source

    go get github.com/nochso/rss
    cd %GOPATH%/src/github.com/nochso/rss
    go run cmd/rssd/main.go

## Usage

    $ rssd -h
    Usage of rssd:
      -grace duration
            HTTP shutdown grace period for existing connections (default 10s)
      -http string
            HTTP listening address (default ":8080")
