#!/bin/bash

cmd=$1

if [ -z "$cmd" ]; then
    echo "please input app name to build"
    exit 1
fi

apps=(
    'mallard2-agent::agent'
    'mallard2-transfer::transfer'
    'mallard2-center::center'
    'mallard2-eventor::eventor'
    'mallard2-alarm::alarm'
    'mallard2-store::store'    
)
isbuild=0

for index in "${apps[@]}"; do 
    KEY="${index%%::*}"
    VALUE="${index##*::}"
    if [ "$KEY" == "$cmd" ]; then
        echo "build $KEY"

        t=$(TZ=Asia/Shanghai date +"%Y-%m-%dT%H:%M:%S%z")
        cat <<EOF > ./$VALUE/build.go
package main

const (
    // BuildTime is auto generated build time
    BuildTime = "$t"
)
EOF

        go build -v -i -o $KEY ./$VALUE/*.go
        ./$KEY -vt

        ver=$(./$KEY -v)

        cfgfile="$KEY""-config.json"
        ./$KEY -c >> $cfgfile

        dir=$KEY
        if [ $KEY == "mallard2-agent" ]; then
            dir="mallard-agent"
        fi

        cat <<EOF > ./$KEY.conf
[program:$KEY]
command = /usr/local/mallard/$dir/$KEY
autostart = true
startsecs = 5
startretries = 5
autorestart = true
user = root
redirect_stderr = true
stdout_logfile = /usr/local/mallard/$dir/var/app.log
stdout_logfile_maxbytes = 200MB
directory = /usr/local/mallard/$dir/
EOF
        echo "tar $KEY-$ver.tar.gz << $KEY $KEY-config.json $KEY.conf"
        tar czf $KEY-$ver.tar.gz $KEY $KEY-config.json $KEY.conf
        rm -rf $KEY $KEY-config.json $KEY.conf

        isbuild=1
    fi
done

if [ $isbuild == 0 ];then
    echo "build nothing for '$cmd'"
    echo "available apps:"
    for index in "${apps[@]}"; do 
        KEY="${index%%::*}"
        printf " - $KEY \n"
    done
fi
