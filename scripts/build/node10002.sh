#! /bin/bash
trap "rm server;kill 0" EXIT

# 单独非后台执行（为了观察日志输出）
cd ..
go build -o server
./server -port 10002