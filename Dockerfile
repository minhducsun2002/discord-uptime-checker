FROM golang:1.24.0-alpine3.20 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /app/check

FROM alpine:3.20.3 as run
WORKDIR /
COPY --from=build /app/check /app/check
ENTRYPOINT /app/check
# mount config as /config.yaml