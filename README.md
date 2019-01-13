sheep库是一个基于grpc的go服务化框架

sheep库的[c++版本](https://github.com/tedcy/sheep_cpp#3221-http-client-in-grpc-server)在这里

- [1 功能](#1-)
- [2 安装](#2-)
- [3 使用](#3-)
    - [3.1 grpc-client](#31-grpc-client)
    - [3.2 server](#32-server)
        - [3.2.1 配置](#321-)
        - [3.2.1 grpc-server](#321-grpc-server)
        - [3.2.2 http-server](#322-http-server)
- [4 extends库使用](#4-extends)
    - [4.1 extends/config      配置文件解析](#41-extendsconfig------)
    - [4.2 extends/log         日志库](#42-extendslog---------)
        - [4.2.1 分不同文件打印](#421-)
        - [4.2.2 不分不同文件打印](#422-)
    - [4.3 extends/coster      调用时长统计](#43-extendscoster------)
    - [4.4 服务端拦截器](#44-)
    - [4.5 flow](#45-flow)

## 1 功能
* grpc原生客户端风格的服务化封装
* 客户端可选的负载均衡策略
* 客户端可选的熔断器
* 服务端可选的多个限流策略

可复用基础库:
* github.com/tedcy/sheep/watcher  独立服务发现封装  
* github.com/tedcy/sheep/breaker  独立熔断器实现  
* github.com/tedcy/sheep/limiter  独立限流器实现  

额外功能extends库(github.com/tedcy/sheep/extends):
* config        读取配置  
* coster        调用时长统计  
* log           日志统计  
* flow          责任链模式封装，适合复杂逻辑请求处理  

注意：目前服务发现只支持etcd组件

## 2 安装
```
go get github.com/tedcy/sheep
```

## 3 使用

### 3.1 grpc-client

客户端目前只支持grpc(http也不需要服务发现)  

具体可参考example/client/client.go

```
//先设置DialConfig，一共4个参数
//EnableBreak
//BalancerType          负载均衡策略，设置client.RespTimeBalancer，则根据调用超时时间反比设置调用权重
//Timeout               调用超时时间
//TargetPath            可被watcher解析的字串，目前只支持etcd，格式为etcd://xxxx:2379,xxxx:2379,xxxx:2379/注册路径
c := &client.DialConfig{}

c.TargetPath = "etcd://172.16.176.38:2379/test1"

conn, err := client.DialContext(context.Background(), c)
//判断err

//接下来使用和原生grpc相同
realConn := pb.NewGreeterClient(conn)

resp, err := realConn.SayHello(context.Background(), &pb.HelloRequest{Name : "name"})
//判断err

fmt.Println("resp: " + resp.Message)
```

### 3.2 server

服务端目前支持http和grpc，两者的配置部分是一样的，使用方式不同

#### 3.2.1 配置

sheep\_server.ServerConfig  
* Addr  
监听地址
* Type
server类型，可选http或grpc  
* WatcherAddrs  
watcher地址，这里是etcd地址列表  
etcd://xxxx:2379,xxxx:2379,xxxx:2379
* WatcherTimeout  
和watcher连接的超时时间，决定了临时节点多久下线
* WatcherPath  
在wathcer上注册路径
* LimiterType  
限流器类型，可选  
limiter.QueueLengthLimiterType，固定请求队列长度限流  
limiter.InvokeTimeLimiterType，根据调用时间限制限流  
* Limit  
限制值  
对QueueLengthLimiterType来说，是请求队列个数  
对InvokeTimeLimiterType来说，是限制的延时毫秒数  

* Interceptors  
自定义拦截器，下面单独说明  
* Opt  
额外的服务端选项，目前只支持grpc, grpc高并发需要带这个选项，否则会因为stream太少报错  
```
config.Opt = &grpc.GrpcServerOpt{
	GrpcOpts: append([]real_grpc.ServerOption{}, real_grpc.MaxConcurrentStreams(10000)),
}
```

#### 3.2.1 grpc-server

具体参考example/server/grpc/server.go

类的实现部分，实现proto生成的方法

```
import (
    pb "google.golang.org/grpc/examples/helloworld/helloworld"
)
type server struct {
}
var gResp *pb.HelloReply = &pb.HelloReply{Message: "Hello"}
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
    return gResp, nil
}
```

使用部分

```
import (
    sheep_server "github.com/tedcy/sheep/server"
    "github.com/tedcy/sheep/server/real_server/grpc"
    real_grpc "google.golang.org/grpc"
)
config := &sheep_server.ServerConfig{}
config.Addr = ":50051"
config.Type = "grpc"
config.Opt = &grpc.GrpcServerOpt{
    GrpcOpts: append([]real_grpc.ServerOption{}, real_grpc.MaxConcurrentStreams(10000)),
}
realS := &server{}
s, err := sheep_server.New(context.Background(), config)
//判断err

//退出关闭服务
defer s.Close()

//获取server的grpc注册器
h, ok := s.GetRegisterHandler().(*real_grpc.Server)
//判断ok

//把realS注册到grpcServer上
pb.RegisterGreeterServer(h, realS)

//服务开启，这里阻塞主协程不会退出
err = s.Serve()
//判断err
```

#### 3.2.2 http-server

参考example/server/http/server.go

服务端结构体需要实现HttpHandlerI的三个方法

**注：http服务端的设计较为复杂，是为了将解析部分剥离出来，让grpc和http服务端代码可以灵活转换，让内网的grpc服务可以随时对外服务**

```
type HttpMap interface{
	Get(key string) string
}
type HttpReq struct {
	req *http.Request
	Headers HttpMap
	QueryStrs HttpMap 
	Body io.Reader
	Path string
}
type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
}
type HttpHandlerI interface{
    //对http请求做解析，解析结果返回req对象，会传给Handler的req
	Decode(httpReq *HttpReq) (req interface{},err error)
    //处理Handler解析的结果，返回resp对象
	Handler(ctx context.Context, req interface{}) (resp interface{}, err error)
    //对Handler返回的resp对象做处理，进行返回
	Encode(resp interface{}, outputErr error, rw ResponseWriter) (err error)
}
```

```
import (
    "github.com/tedcy/sheep/server/real_server/http"
    "io/ioutil"
)
type server struct {
}
func (s *server) Decode(httpReq *http.HttpReq) (req interface{},err error) {
	b, err := ioutil.ReadAll(httpReq.Body)
	if err != nil {
		return
	}
	req = b
	return
}
func (s *server) Encode(resp interface{}, outputErr error, rw http.ResponseWriter) error {
	if outputErr != nil {
		rw.Write([]byte(outputErr.Error()))
		return outputErr
	}
	_, err := rw.Write(resp.([]byte))
	return err
}
func (s *server) Handler(ctx context.Context, req interface{}) (resp interface{}, err error) {
	resp = req
	return
}
```

```
import (
    sheep_server "github.com/tedcy/sheep/server"
)
config := &sheep_server.ServerConfig{}
config.Addr = "127.0.0.1:80"
config.Type = "http"
realS := &server{}

s, err := sheep_server.New(context.Background(), config)
//判断err

//将服务端实现注册到http路径上去，method部分大小写兼容
err = s.Register("GET:/test", realS)
//判断err

//服务开启，这里阻塞主协程不会退出
err := s.Serve()
//判断err
```

## 4 extends库使用

### 4.1 extends/config      配置文件解析

目前是对toml库的封装，不能选择其他格式

```
configs := map[string]interface{} {
    "conf/project.toml" : &Project,
    "conf/redis.toml" : &Redis,
}
for file, cfg := range configs {
    err := config.Read(cfg, file)
    //判断err
}
```

PS: toml格式非常简单，很适合写配置文件

基本规则如下，我没用过这个以外的语法

* k = v, v如果是字符串加""
* 某个结构体，缩进2个空格以后[结构体变量名]，换行再缩进2个空格

```
ClickBaseUrl = ""
WinBaseUrl = ""
  [ServerConfig]
    Addr = ":8235"
    Type = "http"
    LimiterType = 2
    Limit = 100
```

### 4.2 extends/log         日志库

日志库优点是支持自定义分文件，缺点是不支持按日志级别打印。


```
log.GetLog().Infoln()
```

GetLog函数中的参数是为了传入LogOption，来进行分不同文件  
如果不传也可以，会自动获取协程变量来选择文件，如果没有设置协程变量会使用默认全局变量

#### 4.2.1 分不同文件打印
我一般配合config库使用  

比如bidding.toml
```
LogCategory = "file"
Ignore = false
OutputFile = "log/bidding"
HeaderFormat = "$L $datetime-ms $gid $file:$line] "
RotateCategory = "size"
RotateValue = 1912602624
RotateSuffixFormat = ".P$pid.$datetime"
```

```
//解析LogOption
var LogBidding = log.DefaultLogOption()
err := config.Read(&LogBidding, "bidding.toml")

//将LogOption绑定在一个LogKey的变量上
type biddingLogKeyT struct{}
var biddingLogKey biddingLogKeyT
log.BindOption(biddingLogKey, LogBidding)

//下面两行一般会放在服务端拦截器中
//逻辑执行前设置协程变量
log.SetGlsDefaultKey(&LogBidding)
//逻辑结束后清理协程变量
defer log.CleanupGlsDefaultKey(&LogBidding)
```

#### 4.2.2 不分不同文件打印

嫌麻烦的话按下面写可以修改默认全局变量进行打印
```
logOpt := log.DefaultLogOption()
logOpt.LogCategory = "file"
lg, err := log.MakeLogger(logOpt)
//判断err
log.SetLog(lg)
```

### 4.3 extends/coster      调用时长统计

```
//调用时长分类变量
var CosterRequest string = "1"

//逻辑执行前通过coster管理器单例创建一个costerOnce对象，此时开始计时
costerOnce := coster.GetInstance().Start()

//逻辑结束后通过coster管理器对CosterRequest结束costerOnce对应的计时，结束计数
defer coster.GetInstance().GetCoster(CosterRequest).End(costerOnce)

//统计最近五分钟平均数
last := time.Now().Add(-time.Minute * 5)
coster.GetInstance().GetCoster(CosterRequest).GetAverage(last)
```

目前每秒进行一次统计，也就是300次统计的平均值，但是如果当时统计值为0，不会计入结果  
* GetAverage是平均值
* GetMost是众数  
按阶段分类，0-9阶段，10-99阶段，100-999阶段最多的阶段，除以这个阶段的阶段最小值抹掉小数点，再乘100，得到最多的值  
例如0,10,100,101,200,众数就是先算最多的阶段，100-999  
在100,101,200中，除以100，为1,1,2，再乘100，100,100,200，得到众数100  
* GetMax是取最大值  

### 4.4 服务端拦截器

为了让设计更灵活

分类日志，字段解析的日志打印，统计延时等功能我都放在了服务端拦截器中

配合extends库实现


```
import (
    sheep_server "github.com/tedcy/sheep/server"
)
//sheep_server的配置文件加上拦截器
var config sheep_server.ServerConfig
config.Interceptors = append([]server_common.ServerInterceptor{}, serverInterceptor)

//参数
//ctx       上下文信息
//req       传入的请求
//handler   被拦截的handler
//resp      传出的resp
//err       报错信息
func serverInterceptor(ctx context.Context, req interface{},
    handler server_common.ServerHandler) (resp interface{}, err error) {
    var logKey interface{}
    //服务端通过ctx.Value("serviceName").(string)获取到路径名称
    //grpc服务端的serviceName是proto生成代码的Desc描述类
    //http服务端的serviceName是自己注册的Api路径
    serviceName := ctx.Value("serviceName").(string)
    switch serviceName {
    case "/proto.BidService/Handler":
        //每个实现包把它自己的LogKey封装成了GetLogKey()
        logKey = bapi.GetLogKey()
    case "GET:/imp":
        logKey = iapi.GetLogKey()
    case "GET:/click":
        logKey = capi.GetLogKey()
    }

    //日志的LogOption
    log.SetGlsDefaultKey(logKey)
    //释放LogOption
    defer log.CleanupGlsDefaultKey(logKey)

    //对请求json序列化进行打印
    breq, _ := json.Marshal(req)
    log.GetLog().Debugln(serviceName, string(breq))

    //执行被拦截的handler
    resp, err = handler(ctx, req)

    //如果出错打印错误信息，否则打印resp
    if err != nil {
        log.GetLog().Errorln(err)
    }else {
        bresp, _ := json.Marshal(resp)
        log.GetLog().Debugln(string(bresp))
    }
    return
}
```

### 4.5 flow

见doc/flow.md
