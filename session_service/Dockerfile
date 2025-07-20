FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev libffi-dev openssl-dev pkgconfig git

WORKDIR /go/src/session

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o session ./cmd/

FROM alpine:latest AS runtime

WORKDIR /app

COPY --from=builder /go/src/session/session /app/

COPY --from=builder /go/src/session/.env /app/

RUN apk add --no-cache openssl

RUN chmod +x ./session

EXPOSE 8080

ENTRYPOINT ["./session"]
