# Build
ARG GO_VERSION=1.20.4

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
    # use the C toolchain that the *target image* requires, and compile for \
    # that arch. Our images are ubuntu-based, so GCC. \
    CC=x86_64-linux-gnu-gcc \
    GOARCH=amd64 \
    CGO_ENABLED=0 \
    go build -v -o /go/bin/app ./cmd/${APP}

#################
# Runtime stage #
#################

# The base image is *only available* for amd64. With the platform=amd64 flag,
# we're not changing anything, only making it explicit.
# Thanks to emulation, this will also run on ARM Macs.
# Advised to move from distroless to the secure base image - https://docs.google.com/document/d/1I-vxsuHlmBlM8JHSDpvOmVMGeQQcbPgb8jH1ELEE9wo/edit#heading=h.1xke9mez8zov
FROM --platform=amd64 gcr.io/snyk-main/ubuntu-20:1.0.1_202304051233

COPY config.*.json /srv/app
COPY --from=builder /go/bin/app /srv/app/

USER snyk

EXPOSE 8080
CMD ["/srv/app/app"]
