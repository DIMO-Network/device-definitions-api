# device-definitions-api

Api for managing device definitions on the DIMO platform.

## Developing locally

**TL;DR**

```bash
cp settings.sample.yaml settings.yaml
docker compose up -d
go run ./cmd/device-definitions-api migrate
go run ./cmd/device-definitions-api
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

### Database ORM

This is using [sqlboiler](https://github.com/volatiletech/sqlboiler). The ORM models are code generated. If the db changes,
you must update the models.

Make sure you have sqlboiler installed:

```bash
go install github.com/volatiletech/sqlboiler/v4@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
```

To generate the models:

```bash
sqlboiler psql --no-tests --wipe
```

_Make sure you're running the docker image (ie. docker compose up)_

If you get a command not found error with sqlboiler, make sure your go install is correct.
[Instructions here](https://jimkang.medium.com/install-go-on-mac-with-homebrew-5fa421fc55f5)

### Adding migrations

Have goose installed, then:

`goose -dir internal/infrastructure/db/migrations create slugs-not-null sql`


