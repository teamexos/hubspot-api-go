BASE_DIR= github.com/teamexos/hubspot-api-go
DETECT_SECRETS_BASELINE= .devops/.secrets.baseline
TEST_CMD= go test -timeout 30s


.PHONY: init
init:
	npm i
	pip3 install detect-secrets
	npx husky add pre-push "make test"
	npx husky add pre-commit "make secrets-scan"

.PHONY: secrets-scan
secrets-scan:
	detect-secrets-hook --baseline ${DETECT_SECRETS_BASELINE} `git diff --cached --name-only`

.PHONY: secrets-update-baseline
secrets-update-baseline:
	detect-secrets scan --update ${DETECT_SECRETS_BASELINE}
	git add ${DETECT_SECRETS_BASELINE}

.PHONY: test
test:
	${TEST_CMD} ${BASE_DIR}/hubspot
