# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/rudeserver ./cmd/rudeserver

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/rudeserver /app/rudeserver
COPY openapi /app/openapi

ENV OPENAPI_SPEC_PATH=/app/openapi/openapi.yaml

EXPOSE 8080

ENTRYPOINT ["/app/rudeserver"]
