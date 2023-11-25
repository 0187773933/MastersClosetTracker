#!/bin/bash
rm -rf .git/
sudo docker system prune -a
./gitPull.sh
./dockerBuild.sh
./dockerRun.sh