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