.PHONY: dev
dev:
	npx wrangler dev --show-interactive-dev-session false --live-reload

.PHONY: build
build:
	go run github.com/syumai/workers/cmd/workers-assets-gen@v0.23.1 -mode=go
	GOOS=js GOARCH=wasm go build -o ./build/app.wasm .

.PHONY: deploy
deploy:
	npx wrangler deploy
