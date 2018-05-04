ip=$1
dir=/root/sheep_client
ssh $ip "killall main"
go build main.go && ssh $ip mkdir -pv $dir && scp main root@$ip:$dir && scp run.sh root@$ip:$dir
