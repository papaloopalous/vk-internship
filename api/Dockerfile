FROM golang:1.24-alpine AS builder
COPY . /go/src/api
WORKDIR /go/src/api
RUN go build -o api ./cmd/

FROM alpine AS runtime
WORKDIR /app
COPY --from=builder /go/src/api/api /app/
COPY --from=builder /go/src/api/config/config.yaml /app/config/
COPY --from=builder /go/src/api/assets /app/assets/
RUN chmod +x ./api
EXPOSE 8080/tcp
ENTRYPOINT ./api