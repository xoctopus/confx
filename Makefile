# go package info
PACKAGES       := $(shell go list ./...)
MODULE_PATH    := $(shell cat go.mod | grep ^module -m 1 | awk '{ print $$2; }' || '')
MODULE_NAME    := $(shell basename $(MODULE_PATH))
TEST_IGNORES   := "_gen.go|.pb.go|_mock.go|_genx_|main.go|hack/testing.go|example/|envx/errors.go"
FORMAT_IGNORES := ".git/,.xgo/,*.pb.go,*_generated.go"

## global env vars
export GOWORK    := off
export HACK_TEST := true

# git info
IS_GIT_REPO := $(shell git rev-parse --is-inside-work-tree >/dev/null 2>&1 && echo 1 || echo 0)
ifeq ($(IS_GIT_REPO),1)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_TAG    := $(shell git describe --tags --abbrev=0)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
else
GIT_COMMIT := ""
GIT_TAG    := ""
GIT_BRANCH := ""
endif
BUILD_AT=$(shell date "+%Y%m%d%H%M%S")

GOTEST=go
GOBUILD=go

# dependencies
DEP_FMT    := $(shell type goimports-reviser > /dev/null 2>&1 && echo $$?)
DEP_LINTER := $(shell type golangci-lint > /dev/null 2>&1 && echo $$?)

show:
	@echo "module:"
	@echo "    path=$(MODULE_PATH)"
	@echo "    module=$(MODULE_NAME)"
	@echo "tools:"
	@echo "    build=$(GOBUILD)"
	@echo "    test=$(GOTEST)"
	@echo "    goimports-reviser=$(shell which goimports-reviser)"
	@echo "    xgo=$(shell which xgo)"
	@echo "    gocyclo=$(shell which gocyclo)"
	@echo "git:"
	@echo "    commit_id=$(GIT_COMMIT)"
	@echo "    tag=$(GIT_TAG)"
	@echo "    branch=$(GIT_BRANCH)"
	@echo "    build_time=$(BUILD_AT)"
	@echo "    name=$(MODULE_NAME)"

# install dependencies
dep:
	@echo "==> installing dependencies"
	@if [ "${DEP_FMT}" != "0" ]; then \
		echo "    goimports-reviser for format sources"; \
		go install github.com/incu6us/goimports-reviser/v3@latest; \
	fi
	@if [ "${DEP_LINTER}" != "0" ]; then \
		echo "\tgolangci-lint for code static checking"; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest ;\
	fi

upgrade-dep:
	@echo "==> upgrading dependencies"
	@echo "    goimports-reviser for format sources"
	@go install github.com/incu6us/goimports-reviser/v3@latest
	@echo "    gocyclo for calculating cyclomatic complexities of functions"
	@go install github.com/gordonklaus/ineffassign@latest
	@echo "    golangci-lint for code static checking"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

update:
	@echo "==> update depended modules to latest"
	@go get -u all
	@go mod tidy

tidy:
	@echo "==> tidy"
	@go mod tidy

test: dep tidy hack_dep_run
	@echo "==> run unit test"
	@$(GOTEST) test ./... -race -failfast -parallel 1 -gcflags="all=-N -l"

cover: dep tidy hack_dep_run
	@echo "==> run unit test with coverage"
	@$(GOTEST) test ./... -failfast -parallel 1 -gcflags="all=-N -l" -covermode=count -coverprofile=cover.out
	@grep -vE $(TEST_IGNORES) cover.out > cover2.out && mv cover2.out cover.out

hack_dep_run:
	@cd hack && (PODMAN_COMPOSE_WARNING_LOGS=false podman compose up -d --remove-orphans || docker compose up -d)

hack_dep_stop:
	@cd hack && (PODMAN_COMPOSE_WARNING_LOGS=false podman compose down -v || docker compose down -v )

ci-cover: cover

view-cover: cover
	@echo "==> run unit test with coverage and view"
	@$(GOBUILD) tool cover -html cover.out

fmt: dep clean
	@echo "==> formating code"
	@goimports-reviser -rm-unused \
		-imports-order 'std,general,company,project' \
		-project-name ${MODULE_PATH} \
		-excludes $(FORMAT_IGNORES) ./...

lint: dep
	@echo "==> linting"
	@echo ">>>golangci-lint"
	@golangci-lint run
	@echo "done"

pre-commit: dep update lint fmt view-cover

clean:
	@find . -name cover.out | xargs rm -rf
	@find . -name .xgo | xargs rm -rf
	@rm -rf build/*
