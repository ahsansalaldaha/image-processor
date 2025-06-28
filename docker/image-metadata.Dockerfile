FROM golang:1.23-alpine

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN go build -o image-metadata ./cmd/image-metadata

CMD ["./image-metadata"]