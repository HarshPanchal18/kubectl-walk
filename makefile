VERSION="$$(yq .spec.version walk.yaml)"

version:
	@echo "$(VERSION)"

build:
	go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk

apply: build
	chmod +x kubectl-walk
	mv kubectl-walk ~/.local/bin

pkg-bin:
	GOOS=linux   GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk-linux-amd64
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk-darwin-amd64
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk-windows-amd64.exe

pkg-tar: pkg-bin
	tar -czf kubectl-walk-linux-amd64.tar.gz kubectl-walk-linux-amd64 LICENSE
	tar -czf kubectl-walk-darwin-amd64.tar.gz kubectl-walk-darwin-amd64 LICENSE
	tar -czf kubectl-walk-darwin-arm64.tar.gz kubectl-walk-darwin-arm64 LICENSE
	zip kubectl-walk-windows-amd64.zip kubectl-walk-windows-amd64.exe LICENSE

sha: pkg-tar
	sha256sum kubectl-walk-*.tar.gz
	sha256sum kubectl-walk-*.zip