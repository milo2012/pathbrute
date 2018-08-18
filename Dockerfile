FROM debian:jessie-slim
RUN apt-get update
RUN apt-get install -y ca-certificates tar
ADD https://github.com/milo2012/pathbrute/releases/download/v0.0.10/pathbrute_0.0.10_linux_amd64.tar.gz /tmp
RUN tar -xf /tmp/pathbrute_0.0.10_linux_amd64.tar.gz --directory /tmp
ADD https://github.com/milo2012/pathbrute/blob/master/pathbrute.sqlite?raw=true /tmp
RUN cp /tmp/pathbrute.sqlite /
RUN cp /tmp/pathbrute.sqlite /home/
RUN mv /tmp/pathbrute /home/
WORKDIR /home/
ENTRYPOINT ["./pathbrute"]

