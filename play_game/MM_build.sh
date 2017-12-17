#! /bin/bash

killall server

cd ~/Ele_Project/clairstormeye/src/github.com/eleme/purchaseMeiTuan/8080
nohup go build server.go &
rm -f nohup.out
nohup ./server &

cd ~/Ele_Project/clairstormeye/src/github.com/eleme/purchaseMeiTuan/8081
nohup go build server.go &
rm -f nohup.out
nohup ./server &

