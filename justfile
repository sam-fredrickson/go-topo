test:
    go test -v ./... -count=1

cover:
    go test -coverprofile=coverage.out ./topo
    go tool cover -html=coverage.out -o coverage.html
    go tool cover -func=coverage.out

lint:
    golangci-lint run --fix
