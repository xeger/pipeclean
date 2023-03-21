default:
	cat data.sql | go run main.go > data2.sql
	diff data.sql data2.sql
