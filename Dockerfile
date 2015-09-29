FROM golang:1.4.2
MAINTAINER Gabe Conradi <gabe.conradi@gmail.com>
RUN go install github.com/tools/godep
COPY . /src
WORKDIR /src
RUN make setup && make


