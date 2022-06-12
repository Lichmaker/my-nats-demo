启动一个集群

```
docker-compose -f nats-cluster.yaml up -d 
```

启动一个客户端，测试集群是否正常启动
```
docker run --network nats --rm -it synadia/nats-box

# 进去之后， 进行测试， 进行订阅
nats sub -s "nats://lichmaker:123456@nats-1:4222" hello-topic

# 再起一个cli，进行消息发送
nats pub -s "nats://lichmaker:123456@nats-1:4222" hello-topic test-my-nats-cluster
```