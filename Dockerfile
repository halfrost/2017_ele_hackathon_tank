FROM docker-hack.ele.me/mirror/golang:1.9.2-alpine3.6
RUN mkdir /data
COPY server/start.sh /data/start.sh
COPY server/server /data/server
RUN chmod 777 /data/start.sh