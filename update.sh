#!/bin/bash
./gitPull.sh
rm -rf .git/
./dockerBuild.sh
./dockerRun.sh
# sudo docker system prune -a