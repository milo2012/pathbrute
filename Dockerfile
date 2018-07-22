FROM debian:jessie-slim
RUN apt-get update
RUN apt-get install -y ca-certificates
ADD pathBrute_linux /
