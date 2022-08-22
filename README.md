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
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
```
2. Run protoc in the root directory
```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/grpc/device_definition.proto
```