default:
	cat input.sql | ./sqlstream scrub

bin: $(find . -type f -name '*.go')
	mkdir -p bin
	env GOOS=darwin GOARCH=amd64 go build -o bin/sqlstream-darwin-amd64
	env GOOS=darwin GOARCH=arm64 go build -o bin/sqlstream-darwin-arm64
	touch bin

benchmark:
	time cat benchmark.sql | ./sqlstream scrub > /dev/null

clean:
	rm -Rf bin

city: city.markov.json
	./sqlstream generate city.markov.json

# TODO: turn the hit-rate check into a command (or do reinforcement learning?)
city.markov.json: city.csv
	cat city.csv | ./sqlstream train words 5 > city.markov.json
	@ALL=$$(cat city.csv | wc -l); \
HITS=$$(cat city.csv | ./sqlstream recognize --confidence=0.5 city.markov.json | wc -l); \
echo "Hit rate: $${HITS}/$${ALL}"
