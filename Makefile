run:
	go run cmd/web/main.go

test:
	go run gotest.tools/gotestsum@latest --hide-summary=skipped ./...
