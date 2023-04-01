default:
	cat testdata/sql/input.sql | ./sqlstream scrub testdata/en-US/models

bin: $(find . -type f -name '*.go')
	mkdir -p bin
	env GOOS=darwin GOARCH=amd64 go build -o bin/sqlstream-darwin-amd64
	env GOOS=darwin GOARCH=arm64 go build -o bin/sqlstream-darwin-arm64
	touch bin

benchmark:
	time cat testdata/sql/benchmark.sql | ./sqlstream scrub testdata/en-US/models > /dev/null

clean:
	rm -Rf bin
	rm testdata/*/models/*

testdata: testdata/en-US/models/city.json

testdata/en-US/models/city.txt: testdata/en-US/training/city.csv
	tail -n+2 testdata/en-US/training/city.csv | ./sqlstream train dict > testdata/en-US/models/city.txt

testdata/en-US/models/city.json: testdata/en-US/models/city.txt
	cat testdata/en-US/models/city.txt | ./sqlstream train markov:words:5 > testdata/en-US/models/city.json
