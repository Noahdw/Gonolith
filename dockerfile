FROM golang:1.22-alpine AS dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gonolith ./cmd/main.go
CMD ["./gonolith"]

FROM golang:1.22-alpine AS prod
WORKDIR /app
COPY . .
RUN go build -o gonolith ./cmd/main.go
CMD ["./gonolith"]