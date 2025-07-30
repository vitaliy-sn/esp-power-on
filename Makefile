APP_NAME := esp-power-on
GO      ?= go

.PHONY: all build clean

all: build

build:
	$(GO) build -o $(APP_NAME) main.go

clean:
	rm -f $(APP_NAME)
