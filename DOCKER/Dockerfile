# docker build -t node:latest -d DOCKER/Dockerfile .
# docker build --no-cache -t node:latest -f DOCKER/Dockerfile .
# docker run -d -p 3000:3000 -p 8841:8841 -p 26656:26656 -p 26657:26657 -p 26660:26660 -v /node/node/:/root/noah/ node:latest

FROM golang:1.12-buster as builder

ENV APP_PATH /home/noah

COPY . ${APP_PATH}

WORKDIR ${APP_PATH}

RUN set -x \
    && make create_vendor \
    && make build

FROM debian:buster
COPY --from=builder /home/noah/build/noah /usr/local/bin/noah

EXPOSE 3000 26656 8841 26657 26660
CMD ["noah", "node"]
STOPSIGNAL SIGTERM

