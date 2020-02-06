FROM golang:alpine as builder
ENV GO111MODULE=on
WORKDIR /usr/src
COPY at3 at3
COPY athelper athelper
WORKDIR /usr/src/at3
RUN CGO_ENABLED=0 GOOS=linux go build -o acexis-ft-test3 main.go

FROM alpine
WORKDIR /usr/src
ENV APP_HOST=testserver
ENV APP_PORT=3000
ENV TRAINING_SERVER="http://training.phuonghau.com"
COPY --from=builder /usr/src/at3/acexis-ft-test3 .
CMD ["./acexis-ft-test3"]