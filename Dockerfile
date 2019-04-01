FROM alpine:latest
ADD bin/k8s-webshell /data/k8s-webshell
ADD configs/inside_finup.crt /data/inside_finup.crt
ADD configs/inside_finup.key /data/inside_finup.key

RUN chmod 755 /data/k8s-webshell

WORKDIR /data

ENTRYPOINT ["./k8s-webshell"]


