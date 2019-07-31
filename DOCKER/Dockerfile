FROM alpine:3.7

ENV PILLARDEV /noah

RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq bash && \
    addgroup pillardev && \
    adduser -S -G pillardev pillardev -h "$PILLARDEV"

USER pillardev

VOLUME [ $PILLARDEV ]

WORKDIR $PILLARDEV

# api port
EXPOSE 8841 46658

ENTRYPOINT ["/usr/bin/noah"]
CMD ["/usr/bin/noah"]
STOPSIGNAL SIGTERM

ARG BINARY=noah
COPY $BINARY /usr/bin/noah

