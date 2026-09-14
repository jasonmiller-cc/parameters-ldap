# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS builder

ARG VERSION=dev

WORKDIR /src

# Cache module downloads before copying source.
COPY go.mod go.sum ./
# parameters-core is a local replace living in a sibling directory outside
# this build context; Docker forbids COPY-ing paths outside the primary
# context, so it's supplied as a named build context instead — see
# --build-context in the Makefile.
COPY --from=parameters-core . /parameters-core
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /parameters-ldap ./cmd/server

# ---- runtime image ----
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /parameters-ldap /parameters-ldap

EXPOSE 8080
ENTRYPOINT ["/parameters-ldap"]
