FROM golang:1.13.4-alpine3.10
WORKDIR /go/src/github.com/amitbet/dicom
COPY . .
RUN apk add --no-cache make git tar zip build-base
RUN make
