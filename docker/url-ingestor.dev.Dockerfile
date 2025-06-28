FROM golang:1.24-alpine

# Install air for hot reloading
RUN go install github.com/air-verse/air@v1.62.0

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the entire project
COPY . ./

# Create air config if it doesn't exist
RUN if [ ! -f .air.toml ]; then \
    air init; \
    fi

# Expose port
EXPOSE 8080

# Use air for hot reloading in development
CMD ["air", "-c", ".air.toml"] 