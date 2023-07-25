#!/bin/bash
./gitPull.sh
./dockerBuild.sh
./dockerRun.sh
# sudo docker system prune -a