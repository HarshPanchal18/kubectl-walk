VERSION=v1.2.2

version:
	@echo "$(VERSION)"

build:
	go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk

apply: build
	chmod +x kubectl-walk
	cp kubectl-walk ~/.local/bin
# or
# 	sudo cp kubectl-walk /usr/local/bin

bin:
	GOOS=linux   GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk
	GOOS=windows GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk.exe

pkg-bin:
	GOOS=linux   GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk
	GOOS=windows GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk.exe

pkg-tar: pkg-bin
	tar -czf kubectl-walk-$(VERSION)-linux-amd64.tar.gz kubectl-walk LICENSE
	tar -czf kubectl-walk-$(VERSION)-darwin-amd64.tar.gz kubectl-walk LICENSE
	tar -czf kubectl-walk-$(VERSION)-darwin-arm64.tar.gz kubectl-walk LICENSE
	tar -czf kubectl-walk-$(VERSION)-windows-arm64.tar.gz kubectl-walk.exe LICENSE

sha:
	sha256sum kubectl-walk-*.tar.gz
	sha256sum kubectl-walk-*.zip

clean:
	rm kubectl-walk-*.tar.gz
	rm kubectl-walk-*.zip
	rm kubectl-walk-v*