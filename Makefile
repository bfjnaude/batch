.PHONY: generate
generate:
	(cd pkg/batch && go generate)

.PHONY: generate
test:
	(cd pkg/batch && go test)

.PHONY: vendor
vendor: 
	go mod tidy && go mod vendor

example:
	go run example.go