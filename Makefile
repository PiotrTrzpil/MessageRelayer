GO=go
GOTEST=$(GO) test
GOCOVER=$(GO) tool cover

.PHONY: test
.PHONY: cover

cover:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCOVER) -html=coverage.out

test:
	$(GOTEST) -v -coverprofile=coverage.out ./...