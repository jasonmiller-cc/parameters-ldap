FROM golang:1.27-alpine AS builder

ARG VERSION=dev

WORKDIR /src

# Cache module downloads before copying source.
COPY go.mod go.sum ./
# parameters-core is provided via replace directive; copy it alongside.
COPY ../parameters-core /parameters-core
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
