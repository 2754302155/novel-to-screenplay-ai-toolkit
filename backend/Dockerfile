FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/novel-to-screenplay-api ./cmd/server

FROM alpine:3.21

WORKDIR /app
COPY --from=build /bin/novel-to-screenplay-api /usr/local/bin/novel-to-screenplay-api

EXPOSE 8080

CMD ["novel-to-screenplay-api"]
