bin: *.go
	mkdir -p bin
	env GOOS=darwin GOARCH=amd64 go build -o bin/sqlstream-darwin-amd64
	env GOOS=darwin GOARCH=arm64 go build -o bin/sqlstream-darwin-x86_64
	touch bin

benchmark:
	cat benchmark.sql | go run *.go > /dev/null

clean:
	rm -Rf bin

diff:
	cat input.sql | go run *.go > output.sql
	diff input.sql output.sql | head -n 1

run:
	cat input.sql | go run *.go

test:
	cat input.sql | go run *.go > output.sql
	mysql -u root -e 'DROP DATABASE IF EXISTS wfp_gp_development; CREATE DATABASE wfp_gp_development;'
	mysql -u root wfp_gp_development < schema.sql
	mysql -u root wfp_gp_development < output.sql
