default:
	cat testdata/sql/input.sql | ./sqlstream scrub testdata/en-US

bin: $(find . -type f -name '*.go')
	mkdir -p bin
	env GOOS=darwin GOARCH=amd64 go build -o bin/sqlstream-darwin-amd64
	env GOOS=darwin GOARCH=arm64 go build -o bin/sqlstream-darwin-arm64
	touch bin

benchmark:
	time cat testdata/sql/benchmark.sql | ./sqlstream scrub testdata/en-US > /dev/null

clean:
	rm -Rf bin
	rm testdata/*/*.json
	rm testdata/*/*.txt

testdata: testdata/en-US/city.json

testdata/en-US/city.txt: testdata/en-US/city.csv
	tail -n+2 testdata/en-US/city.csv | ./sqlstream train dict > testdata/en-US/city.txt

testdata/en-US/city.json: testdata/en-US/city.txt
	cat testdata/en-US/city.csv | ./sqlstream train markov:words:5 > testdata/en-US/city.json
