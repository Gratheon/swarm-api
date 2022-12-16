FROM alpine:3.7

COPY swarm-api /app/swarm-api
COPY config /app/config

RUN chmod +x /app/swarm-api

USER nobody

EXPOSE 60002

WORKDIR /app
CMD /app/swarm-api
