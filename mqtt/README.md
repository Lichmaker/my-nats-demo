先build出一个镜像
```
docker build . -t mynatsdemo
```

启动
```
docker run --name nats-demo-1 -d -p 4222:4222 -p 8222:8222 -p 6222:6222 -p 1883:1883  mynatsdemo 
```