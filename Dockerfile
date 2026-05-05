FROM golang:1.26 AS builder
WORKDIR /app
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -v -o server

FROM alpine
WORKDIR /app
ENV PORT=8000
ENV GIN_MODE=release
COPY --from=builder /app/server .
COPY --from=builder /app/fonts fonts
COPY --from=builder /app/public public
EXPOSE 8000
CMD ["/app/server"]
