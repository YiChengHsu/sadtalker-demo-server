# Start from the build stage
FROM golang:1.19.11-alpine3.18 AS build-env

# go-ethreum requires gcc
RUN apk add --no-cache build-base linux-headers

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o api-server

# Start from the runtime stage
FROM alpine:3.18

WORKDIR /usr/local

COPY --from=build-env /app/api-server .
COPY --from=build-env /app/static /usr/local/static

CMD ["/usr/local/api-server"]