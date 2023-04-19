.PHONY: benchmark bin clean default test

default:
	cat data/sql/input.sql | ./pipeclean -m mysql -p data/policy.json -x data/sql/schema.sql scrub data/models

bin: bin/pipeclean-darwin-amd64 bin/pipeclean-darwin-arm64 bin/pipeclean-linux-amd64 bin/pipeclean-linux-arm64

bin/pipeclean-darwin-amd64:
	env GOOS=darwin GOARCH=amd64 go build -o bin/pipeclean-darwin-amd64

bin/pipeclean-darwin-arm64:
	env GOOS=darwin GOARCH=amd64 go build -o bin/pipeclean-darwin-arm64

bin/pipeclean-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o bin/pipeclean-linux-amd64

bin/pipeclean-linux-arm64:
	env GOOS=linux GOARCH=amd64 go build -o bin/pipeclean-linux-arm64

benchmark:
	time cat data/sql/benchmark.sql | ./pipeclean -m mysql scrub data/models > data/sql/benchmark-output.sql

clean:
	cd bin ; rm -Rf `git check-ignore *`
	cd data/models ; rm -Rf `git check-ignore *`

test:
	go test ./...

data: data/models/city.markov.json data/models/givenName.markov.json data/models/sn.markov.json data/models/streetName.markov.json

data/models/city.markov.json: Makefile data/training/city.csv
	tail -n+2 data/training/city.csv | ./pipeclean train markov:words:5 > data/models/city.markov.json

data/models/givenName.markov.json: Makefile data/training/givenName.csv
	tail -n+2 data/training/givenName.csv | ./pipeclean train markov:words:5 > data/models/givenName.markov.json

data/models/sn.markov.json: Makefile data/training/sn.csv
	tail -n+2 data/training/sn.csv | ./pipeclean train markov:words:5 > data/models/sn.markov.json

data/models/streetName.markov.json: Makefile data/training/streetName.csv
	tail -n+2 data/training/streetName.csv | ./pipeclean train markov:words:5 > data/models/streetName.markov.json
