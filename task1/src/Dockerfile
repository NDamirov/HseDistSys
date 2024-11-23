FROM ubuntu:22.04

RUN apt-get update && \
    apt-get install -y \
        build-essential \
        iptables \
        cmake

COPY . /app
WORKDIR /app

RUN cmake -B build && \
    cmake --build build

WORKDIR /app/build