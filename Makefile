default:
	cat data/sql/input.sql | ./sqlstream scrub data/models

bin: $(find . -type f -name '*.go')
	mkdir -p bin
	env GOOS=darwin GOARCH=amd64 go build -o bin/sqlstream-darwin-amd64
	env GOOS=darwin GOARCH=arm64 go build -o bin/sqlstream-darwin-arm64
	touch bin

benchmark:
	time cat data/sql/benchmark.sql | ./sqlstream scrub data/models > data/sql/benchmark-output.sql

clean:
	rm -Rf bin
	rm data/models/*

data: data/models/city.json data/models/givenName.json data/models/sn.json data/models/streetName.json

data/models/city.txt: data/training/city.csv
	tail -n+2 data/training/city.csv | ./sqlstream train dict > data/models/city.txt

data/models/city.json: Makefile data/models/city.txt
	cat data/models/city.txt | ./sqlstream train markov:words:5 > data/models/city.json

data/models/givenName.json: Makefile data/training/givenName.csv
	tail -n+2 data/training/givenName.csv | ./sqlstream train markov:words:3 > data/models/givenName.json

data/models/sn.json: Makefile data/training/sn.csv
	tail -n+2 data/training/sn.csv | ./sqlstream train markov:words:3 > data/models/sn.json

data/models/streetName.json: Makefile data/training/streetName.csv
	tail -n+2 data/training/streetName.csv | ./sqlstream train markov:words:5 > data/models/streetName.json
