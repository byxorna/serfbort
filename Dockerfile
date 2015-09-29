FROM golang:1.4.2
MAINTAINER Gabe Conradi <gabe.conradi@gmail.com>
COPY . /go/src/github.com/byxorna/serfbort
WORKDIR /go/src/github.com/byxorna/serfbort
RUN make setup && make
ENTRYPOINT ["./serfbort"]

