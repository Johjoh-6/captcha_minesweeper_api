# syntax=docker/dockerfile:1.4
FROM --platform=linux/amd64 debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

ADD bin/linux_amd64/api /usr/bin/catpcha-sweeper

CMD ["catpcha-sweeper"]
