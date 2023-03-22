default:
	cat input.sql | go run *.go

diff:
	cat input.sql | go run *.go > output.sql
	diff input.sql output.sql | head -n 1

test:
	cat input.sql | go run *.go > output.sql
	mysql -u root -e 'DROP DATABASE IF EXISTS wfp_gp_development; CREATE DATABASE wfp_gp_development;'
	mysql -u root wfp_gp_development < schema.sql
	mysql -u root wfp_gp_development < output.sql
