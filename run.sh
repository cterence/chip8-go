#!/bin/sh
xhost +local:docker
docker run -e DISPLAY=$DISPLAY -v /tmp/.X11-unix:/tmp/.X11-unix -v ./roms:/app/roms chip8-go $1
