#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o ../../server ../../main2.go
cd ../../
./server -port=10001 &
./server -port=10002 &
./server -port=10003 -api &

sleep 2
echo ">>> start test"
curl "http://localhost:8000/api/?key=张三" &
curl "http://localhost:8000/api?key=张三" &
curl "http://localhost:8000/api?key=张三" &

# 服务一直启动，剩下的可以自己手动进行测试；http_test2.sh 提供全自动测试
wait