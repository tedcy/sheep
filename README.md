#sheep

```mermaid
graph LR
Balancer-->client
BreakerI-->client
WeighterI-->client
WatcherI[WatcherI = Reslover]-->Balancer
WeighterBalancerI[WeighterBalancerI = LBPolicy]-->Balancer
```
