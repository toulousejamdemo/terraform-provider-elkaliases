GO_FILES?=$$(find . -name '*.go')
TEST_FOLDER:=test

build: fmtcheck
	go install

test: fmtcheck
	go test $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: fmtcheck
	@sh -c "$(TEST_FOLDER)/testacc.sh"

vet:
	@echo "go vet ."
	@go vet $$(go list ./...) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GO_FILES)

fmtcheck:
	@sh -c "$(TEST_FOLDER)/gofmtcheck.sh"

.PHONY: build test testacc vet fmt fmtcheck
