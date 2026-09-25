format-html:
	npx prettier --parser html $(FILE)

test:
	go test ./...

tidy:
	go mod tidy
