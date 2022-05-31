FROM golang:1.17-alpine3.16 as builder

ADD . /app
WORKDIR /app
RUN go build -o ./s3-downloader/s3-downloader ./s3-downloader/main.go

FROM alpine:3.16
COPY --from=builder /app/s3-downloader/s3-downloader /bin/s3-downloader
