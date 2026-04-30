.PHONY: build test dashboard release

dashboard:
	cd dashboard && npm run build
	rm -rf cli/dashboard/static server/dashboard/static
	mkdir -p cli/dashboard/static server/dashboard/static
	cp -r dashboard/dist/* cli/dashboard/static/
	cp -r dashboard/dist/* server/dashboard/static/

build: dashboard
	cd server && go build -o ../bin/pomelo-hook-server .
	cd cli && go build -o ../bin/pomelo-hook .

test:
	cd server && go test ./...
	cd cli && go test ./...
	cd dashboard && npm test

release: dashboard
	mkdir -p bin
	cd cli && GOOS=linux   GOARCH=amd64 go build -o ../bin/pomelo-hook-linux-amd64   .
	cd cli && GOOS=linux   GOARCH=arm64 go build -o ../bin/pomelo-hook-linux-arm64   .
	cd cli && GOOS=darwin  GOARCH=amd64 go build -o ../bin/pomelo-hook-darwin-amd64  .
	cd cli && GOOS=darwin  GOARCH=arm64 go build -o ../bin/pomelo-hook-darwin-arm64  .
	cd cli && GOOS=windows GOARCH=amd64 go build -o ../bin/pomelo-hook-windows-amd64.exe .
