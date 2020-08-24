BASE_DIR= github.com/teamexos/hubspot-api-go
BUILD_CMD= env GOOS=linux go build -ldflags="-s -w" -o
TEST_CMD= go test -timeout 30s

.PHONY: build clean deploy

build:
	dep ensure -v

clean:
	rm -rf ./bin ./vendor Gopkg.lock

test:
	${TEST_CMD} ${BASE_DIR}/hubspot
