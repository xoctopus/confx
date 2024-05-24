PACKAGES=`go list ./... | grep -E -v 'example|proto'`
FORMAT_FILES=`find . -type f -name '*.go' | grep -E -v '_generated.go|.pb.go'`
XGO_OK=$(shell type xgo > /dev/null 2>&1 && echo $$?)

xgo:
	@if [ "${XGO_OK}" != "0" ]; then \
		echo "installing xgo for unit test"; \
		go install github.com/xhd2015/xgo/cmd/xgo@latest; \
	fi

tidy:
	go mod tidy

cover: xgo tidy
	xgo test -failfast ${PACKAGES} -coverprofile=cover.out -covermode=count


test: tidy
	xgo test -race -failfast ${PACKAGES}

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
