FROM golang:1.25-alpine AS go-builder

ARG GOPROXY
ENV GOPROXY=$GOPROXY

COPY go.mod go.sum /app/
COPY src /app/src

WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    go build -ldflags="-s -w" -o main ./src/*.go

RUN --mount=type=cache,target=/var/cache/apk \
    apk add --no-cache upx

RUN upx --lzma /app/main

FROM alpine:latest AS runtime

COPY --from=go-builder /app/main /app/main

WORKDIR /app

RUN chmod +x ./main

EXPOSE 8000

CMD ["/app/main"]