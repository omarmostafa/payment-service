FROM golang:1.22-alpine AS base

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/air-verse/air@latest

EXPOSE 8080

CMD ["air"]
