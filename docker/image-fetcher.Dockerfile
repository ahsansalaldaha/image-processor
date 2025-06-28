FROM golang:1.24-alpine

# Install air for hot reloading
RUN go install github.com/air-verse/air@v1.62.0

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN go build -o image-fetcher ./cmd/image-fetcher

CMD ["./image-fetcher"]