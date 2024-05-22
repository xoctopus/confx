TEST_PACKAGES=`go list ./... | grep -E -v 'example|proto'`
FORMAT_FILES=`find . -type f -name '*.go' | grep -E -v '_generated.go|.pb.go'`

tidy:
	go mod tidy
cover: tidy
	go test -race -failfast -parallel 1 -gcflags="all=-N -l" ${TEST_PACKAGES} -covermode=atomic -coverprofile cover.out
test: tidy
	go test -race -failfast -parallel 1 -gcflags="all=-N -l" ${TEST_PACKAGES}

report:
	@echo ">>>static checking"
	@go vet ./...
	@echo "done\n"
	@echo ">>>detecting ineffectual assignments"
	@ineffassign ./...
	@echo "done\n"
	@echo ">>>detecting icyclomatic complexities over 10 and average"
	@gocyclo -over 10 -avg -ignore '_test|vendor' . || true
	@echo "done\n"

