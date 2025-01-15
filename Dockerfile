ARG GO_VERSION=1.23

###############
# Build stage #
###############
FROM golang:${GO_VERSION}-bullseye as builder
ARG APP
WORKDIR /go/src/${APP}


# Add go module files
COPY go.mod go.sum ./

# Download and cache dependencies in a dedicated layer.
RUN go mod download

# Add source code & build
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    go build -v -o /go/bin/app ./cmd/vu-api/main.go
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    go build -v -o /go/bin/scraper ./cmd/vu-scraper/main.go

#################
# Runtime stage #
#################

# The base image is *only available* for amd64. With the platform=amd64 flag,
# we're not changing anything, only making it explicit.
# Thanks to emulation, this will also run on ARM Macs.
# Advised to move from distroless to the secure base image - https://docs.google.com/document/d/1I-vxsuHlmBlM8JHSDpvOmVMGeQQcbPgb8jH1ELEE9wo/edit#heading=h.1xke9mez8zov
FROM --platform=amd64 gcr.io/snyk-main/ubuntu-20:2.4.0_202501141014

COPY config.*.json /
COPY --from=builder /go/bin/app /
COPY --from=builder /go/bin/scraper /

USER snyk

EXPOSE 8080
CMD ["/app"]
