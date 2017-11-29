#!/bin/bash

cid1=`docker run -it -d $1 /data/start.sh`
cid2=`docker run -it -d $1 /data/start.sh`
docker run -e MAPID=2  --link=${cid1}:red --link=${cid2}:blue docker-hack.ele.me/jiangang.lan/tankengine:10 /data/start.sh

docker stop $cid1 $cid2
