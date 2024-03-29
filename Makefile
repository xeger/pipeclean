.PHONY: benchmark bin clean default test
DATA=testdata

# quick smoke test of learn + verify (tests scrub as a side effect)
default:
	@cat $(DATA)/sql/data.sql | ./pipeclean learn -c $(DATA)/config.json -r -x $(DATA)/sql/schema.sql $(DATA)/models
	@cat $(DATA)/sql/data.sql | ./pipeclean verify -c $(DATA)/config.json -x $(DATA)/sql/schema.sql $(DATA)/models

bin: bin/pipeclean-darwin-amd64 bin/pipeclean-darwin-arm64 bin/pipeclean-linux-amd64 bin/pipeclean-linux-arm64

bin/pipeclean-darwin-amd64:
	env GOOS=darwin GOARCH=amd64 go build -o bin/pipeclean-darwin-amd64

bin/pipeclean-darwin-arm64:
	env GOOS=darwin GOARCH=arm64 go build -o bin/pipeclean-darwin-arm64

bin/pipeclean-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o bin/pipeclean-linux-amd64

bin/pipeclean-linux-arm64:
	env GOOS=linux GOARCH=arm64 go build -o bin/pipeclean-linux-arm64

benchmark:
	time cat $(DATA)/sql/benchmark.sql | ./pipeclean learn -c $(DATA)/config.json -r -x $(DATA)/sql/schema.sql $(DATA)/models
	time cat $(DATA)/sql/benchmark.sql | ./pipeclean scrub -c $(DATA)/config.json -k -x $(DATA)/sql/schema.sql $(DATA)/models > $(DATA)/sql/benchmark-output.sql

clean:
	cd bin ; rm -Rf `git check-ignore *`
	cd $(DATA)/models ; rm -Rf `git check-ignore *`

test:
	go test ./...

testdata: $(DATA)/models/city.markov.json $(DATA)/models/givenName.markov.json $(DATA)/models/sn.markov.json $(DATA)/models/streetName.markov.json

$(DATA)/models/city.markov.json: Makefile $(DATA)/training/city.csv
	tail -n+2 $(DATA)/training/city.csv | ./pipeclean train markov:words:5 > $(DATA)/models/city.markov.json

$(DATA)/models/givenName.markov.json: Makefile $(DATA)/training/givenName.csv
	tail -n+2 $(DATA)/training/givenName.csv | ./pipeclean train markov:words:5 > $(DATA)/models/givenName.markov.json

$(DATA)/models/sn.markov.json: Makefile $(DATA)/training/sn.csv
	tail -n+2 $(DATA)/training/sn.csv | ./pipeclean train markov:words:5 > $(DATA)/models/sn.markov.json

$(DATA)/models/streetName.markov.json: Makefile $(DATA)/training/streetName.csv
	tail -n+2 $(DATA)/training/streetName.csv | ./pipeclean train markov:words:5 > $(DATA)/models/streetName.markov.json
