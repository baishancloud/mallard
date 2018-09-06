#!/bin/bash

cmd=$1
toType=$2

if [ -z "$cmd" ]; then
    echo "please input app name to build"
    exit 1
fi

if [ -z "$toType" ]; then
    toType="tar.gz"
    echo "set packing to tar.gz"
fi    


apps=(
    'mallard2-agent::agent'
    'mallard2-transfer::transfer'
    'mallard2-center::center'
    'mallard2-eventor::eventor'
    'mallard2-alarm::alarm'
    'mallard2-store::store'    
    'mallard2-judge::judge'
)
isbuild=0

rpmBuild(){
    KEY=$1
    ver=$2
    dir=$3
    fix=$4
    arch=$5
    if [ $arch == "7" ]; then
        arch="el7"
    else
        arch="el6"
    fi
    rpmEl6File=$KEY-$ver.$arch.spec
    echo "new rpm spec : $rpmEl6File"
    cp build.spec $rpmEl6File
    sed -i "s/\[appName\]/$KEY/g" $rpmEl6File
    sed -i "s/\[version\]/$ver/g" $rpmEl6File
    sed -i "s/\[fixversion\]/$fix/g" $rpmEl6File
    sed -i "s/\[arch\]/$arch/g" $rpmEl6File
    sed -i "s/\[destDir\]/$dir/g" $rpmEl6File
    sed -i "s#\[srcDir\]#`pwd`#g" $rpmEl6File
    rpmbuild -bb $rpmEl6File
    rm -rf $rpmEl6File
}

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

        go build -v -i -ldflags "-s -w" -o $KEY ./$VALUE/*.go
        if [ ! -f "$KEY" ]; then
            echo "build fail"
            exit 1
        fi
        ./$KEY -vt

        if [ $toType == "exe" ]; then
            echo "only build binary $KEY"
            exit 1
        fi

        ver=$(./$KEY -v)

        cfgfile="$KEY""-config.json"
        ./$KEY -dc > $cfgfile

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
stdout_logfile_maxbytes = 1000MB
directory = /usr/local/mallard/$dir/
EOF
        if [ $toType == "tar.gz" ]; then
            mkdir -p ./var
            echo "this is log directory" >> ./var/readme.txt
            mkdir -p ./datalogs
            echo "this is cache data directory" >> ./datalogs/readme.txt
            echo "tar $KEY-$ver.tar.gz << $KEY $KEY-config.json $KEY.conf var/ datalogs/"
            tar czf $KEY-$ver.tar.gz $KEY $KEY-config.json $KEY.conf var/ datalogs/
            rm -rf $KEY $KEY-config.json $KEY.conf var/ datalogs/
        fi
        if [ $toType == "rpm" ]; then

            mkdir -p rpms

             # set real config file
            realCfgDir=$MALLARD_CFG_PATH
            if [ -z "$realCfgDir" ]; then
                cp $KEY-config.json config.json
            else
                cp $realCfgDir/$KEY-config.json config.json
                echo "use real-config $realCfgDir/$KEY-config.json"
            fi 

            rpmLockFile=$KEY-$ver.rpm.log
            rpmFixNumber="1"
            if [ -f "$rpmLockFile" ]; then
                number=`cat $rpmLockFile`
                let number+=1
                rpmFixNumber=$number
            fi
            echo "use rpm fixnumber $rpmFixNumber"

            # run rpm build
            rpmBuild $KEY $ver $dir $rpmFixNumber "6"
            rpmBuild $KEY $ver $dir $rpmFixNumber "7"
            rm -rf $KEY $KEY-config.json $KEY.conf config.json
            echo $rpmFixNumber > $rpmLockFile
        fi

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
