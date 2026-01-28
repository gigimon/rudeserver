# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/rudeserver ./cmd/rudeserver

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=build /out/rudeserver /rudeserver

EXPOSE 8080

ENTRYPOINT ["/rudeserver"]

