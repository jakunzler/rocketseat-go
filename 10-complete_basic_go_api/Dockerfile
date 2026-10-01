FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/buildinfo.Version=${VERSION} -X github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/buildinfo.Commit=${COMMIT}" \
    -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/api /api

USER nonroot:nonroot
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD ["/api", "healthcheck"]

ENTRYPOINT ["/api"]
