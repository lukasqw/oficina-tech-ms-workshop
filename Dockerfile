FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api cmd/api/main.go

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/api .

EXPOSE 8083

CMD ["./api"]
