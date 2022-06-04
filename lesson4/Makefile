BINARY_NAME=main
# Makefile
build:
# GOARCH=amd64 GOOS=darwin go build -o ${BINARY_NAME}-darwin main.go
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux main.go
# GOARCH=amd64 GOOS=window go build -o ${BINARY_NAME}-windows main.go

run:
	go run cmd/main.go
#	./${BINARY_NAME}

# build_and_run: build run

clean:
# rm ${BINARY_NAME}-darwin
	rm ${BINARY_NAME}-linux
# rm ${BINARY_NAME}-windows

test:
	go test internal/findfiles/

check:
	go test internal/findfiles/

#test_coverage:
# go test ./... -coverprofile=coverage.out

# dep:
# go mod download

vet:
	go vet

lint:
	golangci-lint run -c golangci-lint.yaml
