#!/bin/bash
APP_NAME="public-analytics-server"
sudo docker rm $APP_NAME -f || echo "failed to remove existing analytics server"

id=$(sudo docker run -dit --restart='always' \
--name $APP_NAME \
-p 9337:9337 \
--mount type=bind,source=/home/morphs/DOCKER_IMAGES/AnalyticsTracker/config.json,target=/app/config.json \
$APP_NAME /app/config.json)
echo "ID = $id"

sudo docker logs -f "$id"