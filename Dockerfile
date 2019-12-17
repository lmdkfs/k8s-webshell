FROM alpine:latest
ADD bin/k8s-webshell /data/k8s-webshell

RUN chmod +x /data/k8s-webshell

WORKDIR /data

ENTRYPOINT ["./k8s-webshell"]