FROM alpine:latest
ADD bin/k8s-webshell /data/k8s-webshell
ADD configs/privkey.pem /data/privkey.pem
ADD configs/fullchain.pem /data/fullchain.pem

RUN chmod 755 /data/k8s-webshell

WORKDIR /data

#ENTRYPOINT ["./k8s-webshell"]


