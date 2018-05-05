ip=172.16.186.217
dir=/root/sheep_server
ssh $ip "killall main"
go build main.go && ssh $ip mkdir -pv $dir && scp main root@$ip:$dir && scp run.sh root@$ip:$dir && ssh $ip $dir/run.sh >/dev/null 2>&1 &
