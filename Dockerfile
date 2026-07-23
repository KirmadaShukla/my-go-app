# Multi-stage production image for my-go-app

FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /out/api \
    ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/api /api

ENV APP_ENV=production \
    HTTP_ADDR=:8080 \
    LOG_LEVEL=info

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/api"]
