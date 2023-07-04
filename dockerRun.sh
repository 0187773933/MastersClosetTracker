#!/bin/bash
APP_NAME="public-mct-db-server"
sudo docker rm -f $APP_NAME || echo ""
id=$(sudo docker run -dit \
--name $APP_NAME \
--mount type=bind,source=/home/morphs/mct/config.json,target=/home/morphs/mct/config.json \
-v /home/morphs/mct/save_files/:/home/morphs/mct/save_files \
-p 5950:5950 \
$APP_NAME config.json)
sudo docker logs -f $id