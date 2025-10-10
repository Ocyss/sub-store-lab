FROM golang:1.25-alpine AS go-builder

COPY go.mod go.sum /app/
COPY src /app/src

WORKDIR /app

RUN go build -o main src/*.go


FROM alpine:latest AS runtime

COPY --from=go-builder /app/main /app/main

WORKDIR /app

RUN chmod +x ./main

EXPOSE 8000

CMD ["/app/main"]