BASE_DIR= github.com/teamexos/hubspot-api-go
TEST_CMD= go test -timeout 30s

.PHONY: build clean deploy



test:
	${TEST_CMD} ${BASE_DIR}/hubspot
