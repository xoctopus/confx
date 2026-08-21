NAME=$(shell basename $$PWD)
FEATURE=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT_RAW=$(shell git rev-parse --short HEAD 2>/dev/null || echo "")
COMMIT_AT=$(shell git log -1 --format=%cd --date=format:%Y%m%d%H%M%S 2>/dev/null || echo "")
BUILD_AT=$(shell date +%Y%m%d%H%M%S)
ifeq ($(shell git status --porcelain 2>/dev/null),)
COMMITID=$(COMMIT_RAW)
else
COMMITID=$(COMMIT_RAW)-dirty
endif

VERSION_PATH=main

LDFLAGS="-s -w -X ${VERSION_PATH}.Name=${NAME} \
	-X ${VERSION_PATH}.Feature=${FEATURE} \
	-X ${VERSION_PATH}.Version=${VERSION} \
	-X ${VERSION_PATH}.CommitID=${COMMITID} \
	-X ${VERSION_PATH}.CommitAt=${COMMIT_AT} \
	-X ${VERSION_PATH}.BuildAt=${BUILD_AT} "

build: clean
	go build -ldflags ${LDFLAGS} -o ${NAME}

run: build
	./example

clean:
	@rm -rf example
