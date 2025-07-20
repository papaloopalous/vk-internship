FROM golang:1.24-alpine AS builder
COPY . /go/src/user
WORKDIR /go/src/user
RUN go build -o user ./cmd/

FROM alpine AS runtime
WORKDIR /app
COPY --from=builder /go/src/user/user /app/
COPY --from=builder /go/src/user/.env /app/
RUN chmod +x ./user
EXPOSE 8080/tcp
ENTRYPOINT ./user