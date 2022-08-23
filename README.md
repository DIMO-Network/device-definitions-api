# device-definitions-api
Api for managing device definitions on the DIMO platform.

## Developing locally

**TL;DR**
```bash
cp settings.sample.yaml settings.yaml
docker compose up -d
go run ./cmd/devices-api migrate
go run ./cmd/devices-api
```

## Generating client and server code
1. Install the protocol compiler plugins for Go using the following commands
```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
```
2. Run protoc in the root directory
```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/grpc/device_definition.proto
```

## Linting

`brew install golangci-lint`

`golangci-lint run`

This should use the settings from `.golangci.yml`, which you can override.

If brew version does not work, download from https://github.com/golangci/golangci-lint/releases (darwin arm64 if M1), then copy to /usr/local/bin and sudo xattr -c golangci-lint