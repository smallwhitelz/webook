FROM ubuntu:20.04
COPY webook /app/webook
WORKDIR /app
CMD ["/app/webook"]