
build: OUTPUT_DIR ?= bin
build: TARGETOS ?= linux
build: TARGETARCH ?= amd64
build:
	@echo "Building..."
	$Q CGO_ENABLED=0 GOOS=${TARGETOS}  GOARCH=${TARGETARCH}  go build -ldflags="-s -w" -o ${OUTPUT_DIR}/ ./ddns_webhook/...
