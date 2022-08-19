FROM golang:1.18 AS build

RUN useradd -u 10001 dimo

WORKDIR /go/src/github.com/DIMO-Network/device-definitions-api/
COPY . /go/src/github.com/DIMO-Network/device-definitions-api/

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOFLAGS=-mod=vendor

RUN ls
RUN go mod tidy
RUN go mod vendor
RUN make install

FROM busybox AS package

LABEL maintainer="DIMO <hello@dimo.zone>"

WORKDIR /

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /go/src/github.com/DIMO-Network/device-definitions-api/target/bin/devices-api .
COPY --from=build /go/src/github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/migrations ./internal/infrastructure/db/migrations

USER dimo

EXPOSE 8080
EXPOSE 8888

CMD /device-definitions-api
