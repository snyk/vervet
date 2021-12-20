# Build
FROM golang:1.17 as builder
ADD . /src
WORKDIR /src
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server

# Run
FROM alpine:3.14
WORKDIR /
COPY --from=builder /src/server .
EXPOSE 8080
CMD ["./server"]
