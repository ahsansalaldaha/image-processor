FROM golang:1.23-alpine

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN go build -o image-fetcher ./cmd/image-fetcher

CMD ["./image-fetcher"]