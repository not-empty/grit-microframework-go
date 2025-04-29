reset
APP_ENV=test go test -v -tags=feature -coverpkg=./... -coverprofile=./tests/coverage/coverage-feature.out ./tests/feature/models
go tool cover -func=./tests/coverage/coverage-feature.out -o ./tests/coverage/coverage-feature.txt
go tool cover -html=./tests/coverage/coverage-feature.out -o ./tests/coverage/coverage-feature.html
sed -i 's|<title>Go Coverage Report</title>|<title>Feature Go Coverage Report</title>|g' ./tests/coverage/coverage-feature.html