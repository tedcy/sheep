etcd集群创建参考

# 创建etcd单点

```
./etcd --advertise-client-urls http://127.0.0.1:2380 --listen-client-urls http://127.0.0.1:2379
```

advertise-client-urls是内部端口，单点没什么用  
listen-client-urls才是对外服务的端口  

# 创建etcd集群

```
git clone https://github.com/coreos/etcd.git
cd bin
mkdir 1
cp etcd 1
mkdir 2
cp etcd 2
mkdir 3
cp etcd 3
```

打开start-cluster.sh，粘贴入以下内容

```
set -e
index=$1
if [[ $index == "" ]];then
    echo ./start-cluster.sh invalid index $index
    exit 1
fi
if (("$index" != 1)) && (("$index" != 2)) && (("$index" != 3));then
    echo ./start-cluster.sh invalid index $index
    exit 1
fi

addrs[1]="127.0.0.1"
addrs[2]="127.0.0.1"
addrs[3]="127.0.0.1"

ports[1]=2380
ports[2]=2381
ports[3]=2382

serve_ports[1]=2379
serve_ports[2]=2378
serve_ports[3]=2377

my_addr=${addrs[$index]}
my_port=${ports[$index]}
my_serve_port=${serve_ports[$index]}


#./etcd -name etcd-$index -debug \
./etcd -name etcd-$index\
    -initial-advertise-peer-urls http://$my_addr:$my_port \
    -listen-peer-urls http://$my_addr:$my_port \
    -listen-client-urls http://$my_addr:$my_serve_port,http://127.0.0.1:$my_serve_port \
    -advertise-client-urls http://$my_addr:$my_serve_port \
    -initial-cluster-token etcd-cluster \
    -initial-cluster etcd-1=http://${addrs[1]}:${ports[1]},etcd-2=http://${addrs[2]}:${ports[2]},etcd-3=http://${addrs[3]}:${ports[3]} \
    -initial-cluster-state new
```

```
cp start-cluster.sh 1
cp start-cluster.sh 2
cp start-cluster.sh 3
分别到1,2,3 目录执行start-cluster.sh
```

注:实际使用时需要自己微调  
* addrs  
etcd服务地址  
* ports  
etcd服务的内部端口（用于raft内部同步）  
* serve\_ports  
etcd服务的外部端口  

# 测试

```
root@ubuntu:/root/etcd/bin# ./etcdctl --endpoints "http://127.0.0.1:2379" set /test ""

root@ubuntu:/root/etcd/bin# ./etcdctl --endpoints "http://127.0.0.1:2379" ls /
/test
root@ubuntu:/root/etcd/bin# ./etcdctl --endpoints "http://127.0.0.1:2379" get /test
```
