# Build
FROM golang:1.20 as builder
ADD . /src
WORKDIR /src
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server

# Run
FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=builder /src/server .
EXPOSE 8080
CMD ["./server"]
