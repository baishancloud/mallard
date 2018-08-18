#!/bin/bash
ts=`date +%s`
ts=$((ts/60))
ts=$((ts*60))
min=`date|awk '{print $4}'|awk -F: '{print $2}'`
((10#$min%2==0))&&status=1||status=0
cmd="!![{\"metric\": \"auto_test\",\"value\": 0,\"fields\": {\"status\":\"$status\"},\"time\": $ts,\"step\": 10,\"endpoint\": \"other-endpoint\"}]"
echo $cmd