.PHONY: build test dashboard

dashboard:
	cd dashboard && npm run build
	rm -rf cli/dashboard/static
	mkdir -p cli/dashboard/static
	cp -r dashboard/dist/* cli/dashboard/static/

build: dashboard
	cd server && go build -o ../bin/pomelo-hook-server .
	cd cli && go build -o ../bin/pomelo-hook .

test:
	cd server && go test ./...
	cd cli && go test ./...
	cd dashboard && npm test
