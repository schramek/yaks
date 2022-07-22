FROM golang:1.17-bullseye

WORKDIR /app

RUN apt-get update && apt-get -y install maven git curl

RUN curl -LO https://github.com/operator-framework/operator-sdk/releases/download/v1.22.2/operator-sdk_linux_amd64
RUN chmod +x operator-sdk_linux_amd64 && mv operator-sdk_linux_amd64 /usr/local/bin/operator-sdk

COPY . .
RUN ls -lha .

RUN make build

RUN ls -lha .
