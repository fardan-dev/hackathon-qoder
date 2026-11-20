# syntax=docker/dockerfile:1
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS build

RUN apk add --no-cache ca-certificates git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/service ./services/${SERVICE}/cmd/server

FROM alpine:3.21 AS runtime
RUN apk add --no-cache ca-certificates wget && adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=build /out/service /app/service
ENTRYPOINT ["/app/service"]
