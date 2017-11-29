#! /bin/bash

mapid=$MAPID
mapfile=/data/map$mapid.txt
echo "map file is $mapfile"
if [ $mapid -eq 1 ];then
 tank_no=4
 tank_speed=1
 shell_speed=2
 tank_HP=1
 tank_score=1
 flag_score=1
 max_round=100
 round_timeout=2000
elif [ $mapid -eq 2 ];then
 tank_no=4
 tank_speed=2
 shell_speed=4
 tank_HP=1
 tank_score=1
 flag_score=1
 max_round=100
 round_timeout=2000
elif [ $mapid -gt 2 ] ;then
 tank_no=4
 tank_speed=1
 shell_speed=2
 tank_HP=1
 tank_score=1
 flag_score=1
 max_round=200
 round_timeout=2000
fi

player1="red:80"
player2="blue:80"
if [ $# -gt 1 ]; then
	player1=$1
	player2=$2
fi
exec java -jar /data/tank-1.0-SNAPSHOT-jar-with-dependencies.jar $mapfile $tank_no $tank_speed $shell_speed $tank_HP $tank_score $flag_score $max_round $round_timeout $player1 $player2 2>&1  | tee /data/logs/engine.log

