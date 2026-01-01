PACKAGES=$(shell go list ./... | grep -E -v 'pb$|testdata|mock|proto|example|testx/internal')
IGNORED=_gen.go|.pb.go|_mock.go|_genx_
MOD=$(shell cat go.mod | grep ^module -m 1 | awk '{ print $$2; }' || '')
MOD_NAME=$(shell basename $(MOD))

GOTEST=go
GOBUILD=go

# dependencies
DEP_FMT=$(shell type goimports-reviser > /dev/null 2>&1 && echo $$?)
DEP_XGO=$(shell type xgo > /dev/null 2>&1 && echo $$?)
DEP_INEFFASSIGN=$(shell type ineffassign > /dev/null 2>&1 && echo $$?)
DEP_GOCYCLO=$(shell type gocyclo > /dev/null 2>&1 && echo $$?)
DEP_LINTER=$(shell type golangci-lint > /dev/null 2>&1 && echo $$?)

# git info
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --tags --abbrev=0)
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
BUILD_AT=$(shell date "+%Y%m%d%H%M%S")

show:
	@echo "packages:"
	@for item in $(PACKAGES); do echo "    $$item"; done
	@echo "module:"
	@echo "    path=$(MOD)"
	@echo "    module=$(MOD_NAME)"
	@echo "tools:"
	@echo "    build=$(GOBUILD)"
	@echo "    test=$(GOTEST)"
	@echo "    goimports-reviser=$(shell which goimports-reviser)"
	@echo "    xgo=$(shell which xgo)"
	@echo "    ineffassign=$(shell which ineffassign)"
	@echo "    gocyclo=$(shell which gocyclo)"
	@echo "git:"
	@echo "    commit_id=$(GIT_COMMIT)"
	@echo "    tag=$(GIT_TAG)"
	@echo "    branch=$(GIT_BRANCH)"
	@echo "    build_time=$(BUILD_AT)"
	@echo "    name=$(MOD_NAME)"

# install dependencies
dep:
	@echo "==> installing dependencies"
	@if [ "${DEP_FMT}" != "0" ]; then \
		echo "    goimports-reviser for format sources"; \
		go install github.com/incu6us/goimports-reviser/v3@latest; \
	fi
	@if [ "${GOTEST}" = "xgo" ] && [ "${DEP_XGO}" != "0" ]; then \
		echo "    xgo for unit test"; \
		go install github.com/xhd2015/xgo/cmd/xgo@latest; \
	fi
	@if [ "${DEP_INEFFASSIGN}" != "0" ]; then \
		echo "    ineffassign for detecting ineffectual assignments"; \
		go install github.com/gordonklaus/ineffassign@latest; \
	fi
	@if [ "${DEP_GOCYCLO}" != "0" ]; then \
		echo "\tgocyclo for calculating cyclomatic complexities of functions"; \
		go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; \
	fi
	@if [ "${DEP_LINTER}" != "0" ]; then \
		echo "\tgolangci-lint for code static checking"; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest ;\
	fi

upgrade-dep:
	@echo "==> upgrading dependencies"
	@echo "    goimports-reviser for format sources"
	@go install github.com/incu6us/goimports-reviser/v3@latest
	@echo "    ineffassign for detecting ineffectual assignments"
	@go install github.com/gordonklaus/ineffassign@latest
	@echo "    gocyclo for calculating cyclomatic complexities of functions"
	@go install github.com/gordonklaus/ineffassign@latest
	@echo "    golangci-lint for code static checking"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@if [ "${GOTEST}" = "xgo" ]; then \
		echo "    xgo for unit test"; \
		go install github.com/xhd2015/xgo/cmd/xgo@latest; \
	fi

update:
	@GOWORK=off go get -u all
	@GOWORK=off go mod tidy

hack_dep_run:
	@cd hack && (PODMAN_COMPOSE_WARNING_LOGS=false podman compose up -d --remove-orphans || docker compose up -d --remove-orphans)

hack_dep_stop:
	@cd hack && (PODMAN_COMPOSE_WARNING_LOGS=false podman compose down -v || docker compose down -v )

tidy:
	@echo "==> tidy"
	@GOWORK=off go mod tidy

test: dep tidy hack_dep_run
	@echo "==> run unit test"
	GOWORK=off $(GOTEST) test ./... -race -failfast -parallel 1 -gcflags="all=-N -l"

cover: dep tidy hack_dep_run
	@echo "==> run unit test with coverage"
	@GOWORK=off $(GOTEST) test ./... -failfast -parallel 1 -gcflags="all=-N -l" -covermode=count -coverprofile=cover.out
	@grep -vE '_gen.go|.pb.go|_mock.go|_genx_|pkg/envx/errors.go|example/|envconf/errors.go' cover.out > cover2.out && mv cover2.out cover.out

ci-cover: cover

view-cover: cover
	@echo "==> run unit test with coverage and view"
	@GOWORK=off $(GOBUILD) tool cover -html cover.out

fmt: dep clean
	@echo "==> formating code"
	@goimports-reviser -rm-unused \
		-imports-order 'std,general,company,project' \
		-project-name ${MOD} \
		-excludes '.git/,.xgo/,*.pb.go,*_generated.go' ./...

lint: dep
	@echo "==> static check"
	@echo ">>>govet"
	@GOWORK=off $(GOBUILD) vet ./...
	@echo "done"
	@echo ">>>golangci-lint"
	@golangci-lint run
	@echo "done"

pre-commit: dep update lint fmt view-cover

clean:
	@find . -name cover.out | xargs rm -rf
	@find . -name .xgo | xargs rm -rf
	@rm -rf build/*
