BASE_DIR= github.com/teamexos/hubspot-api-go
TEST_CMD= go test -timeout 30s

.PHONY: build clean deploy

build:
	dep ensure -v

clean:
	rm -rf ./bin ./vendor Gopkg.lock

test:
	${TEST_CMD} ${BASE_DIR}/hubspot
