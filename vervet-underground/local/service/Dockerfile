# Build
FROM golang:1.17 as builder
WORKDIR /src

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o server

# Run
FROM alpine:3.14
WORKDIR /
COPY --from=builder /src/server .
EXPOSE 8080
CMD ["./server"]
