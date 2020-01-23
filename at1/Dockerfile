FROM golang:alpine as builder
ENV GO111MODULE=on
WORKDIR /usr/src
COPY at1 at1
COPY athelper athelper
WORKDIR /usr/src/at1
RUN CGO_ENABLED=0 GOOS=linux go build -o acexis-ft-test1 cmd/*.go

FROM alpine
WORKDIR /usr/src
COPY --from=builder /usr/src/at1/acexis-ft-test1 .
CMD ["./acexis-ft-test1", "--host", "testserver"]
