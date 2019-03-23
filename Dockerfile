FROM alpine:latest
ADD bin/k8s-webshell /data/k8s-webshell

RUN chmod 755 /data/k8s-webshell

WORKDIR /data

#ENTRYPOINT ["./k8s-webshell"]


