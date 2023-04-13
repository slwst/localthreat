FROM golang:1.20-alpine

RUN apk add --no-cache build-base git 
RUN go install github.com/pilu/fresh@latest

WORKDIR /usr/local/src
ADD . /usr/local/src

ENV GO111MODULE on
EXPOSE 5000

CMD ["fresh"]