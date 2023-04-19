.PHONY: benchmark bin clean default test
DATA=testdata

default:
	cat $(DATA)/sql/input.sql | ./pipeclean -c $(DATA)/config.json -m mysql -x $(DATA)/sql/schema.sql scrub $(DATA)/models

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

testdata: $(DATA)/models/city.markov.json $(DATA)/models/givenName.markov.json $(DATA)/models/sn.markov.json $(DATA)/models/streetName.markov.json

$(DATA)/models/city.markov.json: Makefile $(DATA)/training/city.csv
	tail -n+2 data/training/city.csv | ./pipeclean train markov:words:5 > $(DATA)/models/city.markov.json

$(DATA)/models/givenName.markov.json: Makefile $(DATA)/training/givenName.csv
	tail -n+2 data/training/givenName.csv | ./pipeclean train markov:words:5 > $(DATA)/models/givenName.markov.json

$(DATA)/models/sn.markov.json: Makefile $(DATA)/training/sn.csv
	tail -n+2 $(DATA)/training/sn.csv | ./pipeclean train markov:words:5 > $(DATA)/models/sn.markov.json

$(DATA)/models/streetName.markov.json: Makefile $(DATA)/training/streetName.csv
	tail -n+2 $(DATA)/training/streetName.csv | ./pipeclean train markov:words:5 > $(DATA)/models/streetName.markov.json
