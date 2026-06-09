.PHONY: build clean test run

VERSION := $(shell date +%Y.%m.%d.%H%M)
NAME := eino-stock

build: frontend backend

frontend:
	cd frontend && npm run build
	rm -rf internal/server/web
	cp -r frontend/dist internal/server/web

backend:
	go build -ldflags "-X main.Version=$(VERSION) -X main.Name=$(NAME)" -o build/$(NAME) ./cmd/$(NAME)/

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.Name=$(NAME)" -o build/$(NAME).exe ./cmd/$(NAME)/

clean:
	rm -rf build/ internal/server/web/

test:
	go test ./... -v -count=1

run:
	go run ./cmd/$(NAME)/ -conf configs/
