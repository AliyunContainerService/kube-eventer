FROM golang:1.12.6 AS build-env
ADD . /src/github.com/AliyunContainerService/kube-eventer
ENV GOPATH /:/src/github.com/AliyunContainerService/kube-eventer/vendor
ENV GO111MODULE on
WORKDIR /src/github.com/AliyunContainerService/kube-eventer
RUN apt-get update -y && apt-get install gcc ca-certificates
RUN make


FROM alpine:3.10

COPY --from=build-env /src/github.com/AliyunContainerService/kube-eventer/kube-eventer /
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENV TZ "Asia/Shanghai"
RUN apk add --no-cache tzdata
COPY deploy/entrypoint.sh /

ENTRYPOINT ["/kube-eventer"]

