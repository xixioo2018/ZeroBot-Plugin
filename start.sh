ps -ef | grep zerobot | grep -v grep | awk '{print $2}' | xargs kill -9

nohup ./zerobot -c config.json >/dev/null 2>log &
