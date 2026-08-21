VERSION=v1.2.3

version:
	@echo "$(VERSION)"

build:
	go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk

apply: build
	chmod +x kubectl-walk
	cp kubectl-walk ~/.local/bin
# OR
# 	sudo cp kubectl-walk /usr/local/bin

bin:
	GOOS=linux   GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk.exe

tar:
	GOOS=linux   GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk
	tar -czf kubectl-walk-$(VERSION)-linux-amd64.tar.gz kubectl-walk LICENSE

	GOOS=darwin  GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk
	tar -czf kubectl-walk-$(VERSION)-darwin-amd64.tar.gz kubectl-walk LICENSE

	GOOS=darwin  GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk
	tar -czf kubectl-walk-$(VERSION)-darwin-arm64.tar.gz kubectl-walk LICENSE

	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -s -w" -o kubectl-walk.exe
	tar -czf kubectl-walk-$(VERSION)-windows-amd64.tar.gz kubectl-walk.exe LICENSE

sha:
	sha256sum kubectl-walk-*.tar.gz

clean:
	rm kubectl-walk*