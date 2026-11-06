# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
	-trimpath \
	-ldflags="-s -w -X main.version=${VERSION}" \
	-o /out/ticketing ./cmd/ticketing

FROM scratch

# TLS roots for remote Postgres / PayHere; time zone data is embedded via time/tzdata.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/ticketing /ticketing

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/ticketing"]
CMD ["serve"]
