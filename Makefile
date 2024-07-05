.DEFAULT_GOAL := build
BINARY_NAME=blutui

build:
	go build \
		-ldflags "-X main.appVersion=$$(git rev-parse --short HEAD)" \
		-o /tmp/${BINARY_NAME} cmd/blutui.go

run: build
	/tmp/${BINARY_NAME}

install: build
	cp /tmp/${BINARY_NAME} $$GOPATH/bin/blutui

clean:
	go clean
	rm /tmp/${BINARY_NAME}
