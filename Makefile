BASE_DIR= github.com/teamexos/hubspot-api-go
TEST_CMD= go test -timeout 30s

.PHONY: test

test:
	${TEST_CMD} ${BASE_DIR}/hubspot
