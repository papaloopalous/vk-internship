FROM golang:1.24-alpine AS builder
COPY . /go/src/listing
WORKDIR /go/src/listing
RUN go build -o listing ./cmd/

FROM alpine AS runtime
WORKDIR /app
COPY --from=builder /go/src/listing/listing /app/
COPY --from=builder /go/src/listing/.env /app/
RUN chmod +x ./listing
EXPOSE 8080/tcp
ENTRYPOINT ./listing