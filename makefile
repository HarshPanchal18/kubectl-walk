VERSION="$$(yq .spec.version .krew.yaml)"

version:
	@echo "$(VERSION)"

build:
	go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk

apply: build
	chmod +x kubectl-walk
	cp kubectl-walk ~/.local/bin
# or
# 	sudo cp kubectl-walk /usr/local/bin

pkg-bin:
	GOOS=linux   GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk-v$(VERSION)-linux-amd64
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk-v$(VERSION)-darwin-amd64
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk-v$(VERSION)-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk-v$(VERSION)-windows-amd64.exe

pkg-tar: pkg-bin
	tar -czf kubectl-walk-v$(VERSION)-linux-amd64.tar.gz kubectl-walk-v$(VERSION)-linux-amd64 LICENSE
	tar -czf kubectl-walk-v$(VERSION)-darwin-amd64.tar.gz kubectl-walk-v$(VERSION)-darwin-amd64 LICENSE
	tar -czf kubectl-walk-v$(VERSION)-darwin-arm64.tar.gz kubectl-walk-v$(VERSION)-darwin-arm64 LICENSE
	zip kubectl-walk-v$(VERSION)-windows-amd64.zip kubectl-walk-v$(VERSION)-windows-amd64.exe LICENSE

sha:
	sha256sum kubectl-walk-*.tar.gz
	sha256sum kubectl-walk-*.zip