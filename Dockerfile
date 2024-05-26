# Build stage for WebAssembly
FROM golang:1.22.2 AS wasm-builder

WORKDIR /build

# Copy go mod and sum files
COPY go.mod ./
RUN go mod download

# Set environment variables for cross-compilation to WebAssembly
ENV CGO_ENABLED=0 GOOS=js GOARCH=wasm

# Copy the source files
COPY . .

# Build the WebAssembly binary
RUN go build -o ./main.wasm ./main.go

# Build stage for the Go web server
FROM golang:1.22.2 AS server-builder

WORKDIR /build

# Copy go mod and sum files
COPY go.mod ./
RUN go mod download

# Set environment variables for native compilation
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Copy the source files
COPY . .

# Build the Go web server
RUN go build -o ./server ./server.go

# Final stage to produce a clean and small image
FROM alpine:latest

# Copy the wasm and server binaries from the build stages
COPY --from=wasm-builder /build/main.wasm /app/main.wasm
COPY --from=server-builder /build/server /app/server

# Copy static files needed for the web server
COPY index.html wasm_exec.js /app/

# Set the working directory
WORKDIR /app

# Command to run the server
CMD ["./server"]
