FROM debian:jessie-slim
RUN apt-get update
RUN apt-get install -y ca-certificates tar
ADD https://github.com/milo2012/pathbrute/releases/download/v0.0.9/pathbrute_0.0.9_linux_amd64.tar.gz /tmp
RUN tar -xf /tmp/pathbrute_0.0.9_linux_amd64.tar.gz --directory /tmp
RUN mv /tmp/pathbrute /home/
WORKDIR /home/
ENTRYPOINT ["./pathbrute"]
