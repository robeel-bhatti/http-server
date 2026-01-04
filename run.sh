#!/usr/bin/env bash

# kill any process using port 8080 just in case
kill $(lsof -t -i:8080)

# write foo.txt file which contains Hello, World! text
echo "Hello, World!" > /internal/tmp/foo.txt

go run main.go &
ServerPID=$!

sleep 2

echo
echo

echo "=====Default Endpoint====="
curl -i http://localhost:8080/

echo
echo

echo "=====User Agent Endpoint====="
curl -i --header "User-Agent: foobar/1.2.3" http://localhost:8080/user-agent

echo
echo

echo "=====Echo Endpoint======"
curl -i http://localhost:8080/echo/abc

echo
echo

echo "=====Read File Endpoint====="
curl -i http://localhost:8080/files/foo

kill "$ServerPID"
