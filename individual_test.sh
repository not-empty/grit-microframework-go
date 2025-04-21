reset
go test -v -tags=unit -coverpkg=./... -coverprofile=./tests/coverage/coverage-unit.out -coverpkg=./app/repository ./tests/unit/repository/base_repository_test.go
go tool cover -func=./tests/coverage/coverage-unit.out -o ./tests/coverage/coverage-unit.txt
go tool cover -html=./tests/coverage/coverage-unit.out -o ./tests/coverage/coverage-unit.html
sed -i 's|<title>handler: Go Coverage Report</title>|<title>Unit Go Coverage Report</title>|g' ./tests/coverage/coverage-unit.html