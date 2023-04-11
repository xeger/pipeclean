.PHONY: benchmark bin clean default test

default:
	cat data/sql/input.sql | ./pipeclean scrub data/models

bin: bin/pipeclean-darwin-amd64 bin/pipeclean-darwin-arm64

bin/pipeclean-darwin-amd64:
	env GOOS=darwin GOARCH=amd64 go build -o bin/pipeclean-darwin-amd64

bin/pipeclean-darwin-arm64:
	env GOOS=darwin GOARCH=amd64 go build -o bin/pipeclean-darwin-arm64

benchmark:
	time cat data/sql/benchmark.sql | ./pipeclean scrub data/models > data/sql/benchmark-output.sql

clean:
	rm -Rf bin
	rm -Rf data/models/*

test:
	go test ./...

data: data/models/city.json data/models/givenName.json data/models/sn.json data/models/streetName.json

data/models/city.txt: data/training/city.csv
	tail -n+2 data/training/city.csv | ./pipeclean train dict > data/models/city.txt

data/models/city.json: Makefile data/models/city.txt
	cat data/models/city.txt | ./pipeclean train markov:words:5 > data/models/city.json

data/models/givenName.json: Makefile data/training/givenName.csv
	tail -n+2 data/training/givenName.csv | ./pipeclean train markov:words:3 > data/models/givenName.json

data/models/sn.json: Makefile data/training/sn.csv
	tail -n+2 data/training/sn.csv | ./pipeclean train markov:words:3 > data/models/sn.json

data/models/streetName.json: Makefile data/training/streetName.csv
	tail -n+2 data/training/streetName.csv | ./pipeclean train markov:words:5 > data/models/streetName.json
