FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
# no dependencies, so this line is currently unnecessary (and crashes)
#RUN go mod download

COPY *.go *.tpl ./
RUN CGO_ENABLED=0 go build -o server .

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8000
CMD ["./server"]