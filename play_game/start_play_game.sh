#! /bin/bash


exec java -cp ./data/client/tank.jar:./data/client/classes ele.me.hackathon.tank.GameEngine ./data/secondweekmap.txt 4 1 2 1 1 1 100 2000 localhost:8080 localhost:8081 2>&1  | tee ./data/logs/engine2.log