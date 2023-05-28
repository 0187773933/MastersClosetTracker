#!/bin/bash
BinaryName="mct"
rm -rf ./bin/

# https://stackoverflow.com/questions/25051623/golang-compile-for-all-platforms-in-windows-7-32-bit
# https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

# declare -a linux_architectures=(
# 	"amd64"
# 	# "arm"
# 	# "arm64"
# )
declare -a darwin_architectures=(
	"amd64"
	# "arm64"
)
declare -a windows_architectures=(
	"amd64"
)
# declare -a web_architectures=(
# 	"wasm"
# )

# for architecture in "${linux_architectures[@]}"
# do
# 	echo "Building Linux: $architecture"
# 	GOOS=linux GOARCH=$architecture go build -o bin/linux/$architecture/$BinaryName
# done

for architecture in "${darwin_architectures[@]}"
do
	echo "Building Darwin: $architecture"
	GOOS=darwin GOARCH=$architecture go build -o bin/darwin/$architecture/$BinaryName
done

for architecture in "${windows_architectures[@]}"
do
	echo "Building Windows: $architecture"
	GOOS=windows GOARCH=$architecture go build -o bin/windows/$architecture/$BinaryName.exe
done

# for architecture in "${web_architectures[@]}"
# do
# 	echo "Building Web: $architecture"
# 	GOOS=js GOARCH=$architecture go build -o bin/windows/$architecture/$BinaryName.wasm
# done


cp ./bin/windows/amd64/mct.exe "/Users/morpheous/Library/CloudStorage/Dropbox/Parking/PERMANENT/Masters Closet Tracker/.files"
cp ./bin/darwin/amd64/mct "/Users/morpheous/Library/CloudStorage/Dropbox/Parking/PERMANENT/Masters Closet Tracker/.files"
cp -r v1/server/cdn "/Users/morpheous/Library/CloudStorage/Dropbox/Parking/PERMANENT/Masters Closet Tracker/.files/v1/server"
cp -r v1/server/html "/Users/morpheous/Library/CloudStorage/Dropbox/Parking/PERMANENT/Masters Closet Tracker/.files/v1/server"
cp -r v1/printer "/Users/morpheous/Library/CloudStorage/Dropbox/Parking/PERMANENT/Masters Closet Tracker/.files/v1/"
