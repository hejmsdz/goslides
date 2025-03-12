FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -v -o server

FROM alpine
RUN apk add --no-cache libc6-compat
WORKDIR /app
VOLUME ["/app/data"]
ENV PORT=8000
ENV GIN_MODE=release
COPY --from=builder /app/server .
COPY --from=builder /app/fonts fonts
COPY --from=builder /app/public public
EXPOSE 8000
CMD ["/app/server"]
