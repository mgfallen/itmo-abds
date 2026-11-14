FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app cmd/app/main.go
RUN go build -o gen cmd/gen/main.go
RUN go build -o load cmd/load/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/gen .
COPY --from=builder /app/load .

EXPOSE 8080

CMD ["./app"]