#!/usr/bin/env bash

# environment variable
Home="$HOME"

mkdir -p bigipct-bin/{bigipct-linux-amd64,bigipct-windows-amd64,bigipct-darwin-amd64}
# build Linux executable package
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o f5-bigipct main.go
mv f5-bigipct bigipct-bin/bigipct-linux-amd64
cd bigipct-bin
tar -zcf bigipct-linux-amd64.tar.gz  ./bigipct-linux-amd64
rm -rf bigipct-linux-amd64

# build Windows executable package
cd ..
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o f5-bigipct.exe main.go
mv f5-bigipct.exe bigipct-bin/bigipct-windows-amd64
cd bigipct-bin
tar -zcf bigipct-windows-amd64.tar.gz ./bigipct-windows-amd64
rm -rf ./bigipct-windows-amd64

# build Mac OS executable package
cd ..
go build -o f5-bigipcts main.go
mv f5-bigipcts bigipct-bin/bigipct-darwin-amd64
cd bigipct-bin
tar -zcf bigipct-darwin-amd64.tar.gz ./bigipct-darwin-amd64
rm -rf ./bigipct-darwin-amd64

cd ..
mv -f bigipct-bin ${Home}/Desktop/