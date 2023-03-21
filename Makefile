default:
	cat data.sql | go run main.go

diff:
	cat data.sql | go run main.go > data2.sql
	diff data.sql data2.sql | head -n 1

test:
	cat data.sql | go run main.go > data2.sql
	diff data.sql data2.sql
