<script src="https://cdnjs.cloudflare.com/ajax/libs/mermaid/8.0.0/mermaid.min.js"></script>
<script>
var config = {
    startOnLoad:true,
    theme: 'forest'
};
mermaid.initialize(config);
window.mermaid.init(undefined, document.querySelectorAll('.language-mermaid'));
</script>

# client实现

<div class="mermaid">
graph LR
Balancer-->client
BreakerNotifier-->client
WeighterNotifier-->client
ConnectionManager[ConnectionManager,impl by grpc]-->grpc.ClientConn
Reslover-->Balancer
LBPolicy-->Balancer
client--grpc.WithBalancer-->grpc.ClientConn
Notifiers--grpc.WithUnaryInterceptor-->grpc.ClientConn
</div>
