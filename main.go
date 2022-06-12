package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

const MY_CLUSTER = "nats://127.0.0.1:4222, nats://127.0.0.1:4223, nats://127.0.0.1:4224"

const MY_TOPIC = "hello-topic"
const MY_TOPIC_REQUEST = "hello-topic-request"

func main() {
	fmt.Println("start demo NOW!")
	natsConnect, err := nats.Connect(MY_CLUSTER, nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(1),
		nats.ReconnectWait(time.Second),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			// Note that this will be invoked for the first asynchronous connect.
			fmt.Println("成功重新连接到nats")
		}),
		nats.UserInfo("lichmaker", "123456"))
	if err != nil {
		panic(err)
	}
	defer natsConnect.Close()
	fmt.Println("connect to nats...")

	natsEncodedConnect, err := nats.NewEncodedConn(natsConnect, nats.JSON_ENCODER)
	if err != nil {
		panic(err)
	}
	defer natsEncodedConnect.Close()
	fmt.Println("build the encoded connects ...")

	// // 测试订阅
	// wg := &sync.WaitGroup{}
	// wg.Add(1)
	// go subscribeDemo(wg, natsEncodedConnect)
	// // 测试发送消息
	// go publishMsgDemo(wg, natsEncodedConnect)

	// 测试订阅一个request
	// wg := &sync.WaitGroup{}
	// wg.Add(1)
	// go subscribeRequestDemo(wg, natsEncodedConnect)
	// go sendToRequestDemo(wg, natsEncodedConnect)

	// 测试订阅一个group
	// wg := &sync.WaitGroup{}
	// wg.Add(1)
	// go publishMsgDemo(wg, natsEncodedConnect)
	// go subscribeQueueDemo(wg, natsEncodedConnect)

	// 测试使用同步消息接受
	// wg := &sync.WaitGroup{}
	// wg.Add(1)
	// go publishMsgDemo(wg, natsEncodedConnect)
	// go subscribeWithSync(wg, natsEncodedConnect)

	// 测试使用同步消息来接受request
	wg := &sync.WaitGroup{}
	wg.Add(1)
	myInbox := nats.NewInbox()
	go subscribeRequestWithInboxDemo(wg, myInbox, natsEncodedConnect)
	go sendToSyncRequestDemo(wg, myInbox, natsEncodedConnect)

	// 大吞吐量的时候， 不会所有消息都马上发到nats server，有可能会到一个buffer里。
	// 按照文档说明，调用flush的时候，会马上处理缓冲区，并且会发一个ping到server， 接收到一个pong之后会返回。
	// 所以调用flush有2个用途，一个是ping/pong， 一个是处理缓冲区。
	// https://docs.nats.io/using-nats/developer/sending/caches
	natsEncodedConnect.Flush()
	fmt.Println("All clear!")

	fmt.Println("sleeping")
	time.Sleep(time.Second * 60)
}

func publishMsgDemo(wg *sync.WaitGroup, conn *nats.EncodedConn) {
	wg.Wait()
	for i := 0; i < 20; i++ {
		fmt.Printf("发送消息 for circle : %d\n", i)
		err := conn.Publish(MY_TOPIC, "hey ， 这是一条消息")
		if err != nil {
			panic(err)
		}
		// time.Sleep(time.Millisecond * 500)
	}
}

func subscribeDemo(wg *sync.WaitGroup, conn *nats.EncodedConn) {
	sb, err := conn.Subscribe(MY_TOPIC, func(mySubject string, value string) {
		fmt.Printf("成功消费： topic - %s , value - %s , timestamp - %d \n", mySubject, value, time.Now().Unix())
		time.Sleep(time.Second)
	})
	if err != nil {
		panic(err)
	}
	wg.Done()

	time.Sleep(time.Second * 5)
	fmt.Println("start drain " + sb.Subject)
	// drain 的用法，就是触发一个事件，当把目前接收到的消息处理完成之后，关闭订阅。
	// sb.Drain()

}

func subscribeRequestDemo(wg *sync.WaitGroup, conn *nats.EncodedConn) {
	_, err := conn.Subscribe(MY_TOPIC_REQUEST, func(subject string, replySubject string, msg string) {
		conn.Publish(replySubject, fmt.Sprintf("从topic收到消息成功 %s 。 现在给 replySubject 发送一个消息。 time %d \n", msg, time.Now().Unix()))
	})
	if err != nil {
		panic(err)
	}
	wg.Done()
}

func sendToRequestDemo(wg *sync.WaitGroup, conn *nats.EncodedConn) {
	wg.Wait()
	var msg string
	for i := 0; i < 10; i++ {
		err := conn.Request(MY_TOPIC_REQUEST, "我是一条request消息", &msg, time.Second*5)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d 收到reply: %s \n", time.Now().Unix(), msg)
		time.Sleep(time.Second)
	}
}

func subscribeQueueDemo(wg *sync.WaitGroup, conn *nats.EncodedConn) {
	queueGroupID := "myGroup"

	go func() {
		_, err := conn.QueueSubscribe(MY_TOPIC, queueGroupID, func(mySubject string, value string) {
			fmt.Printf("成功消费 [%d] ： topic - %s , value - %s , timestamp - %d \n", 1, mySubject, value, time.Now().Unix())
			// time.Sleep(time.Second)
		})
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		_, err := conn.QueueSubscribe(MY_TOPIC, queueGroupID, func(mySubject string, value string) {
			fmt.Printf("成功消费 [%d] ： topic - %s , value - %s , timestamp - %d \n", 2, mySubject, value, time.Now().Unix())
			// time.Sleep(time.Second)
		})
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		_, err := conn.QueueSubscribe(MY_TOPIC, queueGroupID, func(mySubject string, value string) {
			fmt.Printf("成功消费 [%d] ： topic - %s , value - %s , timestamp - %d \n", 3, mySubject, value, time.Now().Unix())
			time.Sleep(time.Second)
		})
		if err != nil {
			panic(err)
		}
	}()
	wg.Done()
}

func subscribeWithSync(wg *sync.WaitGroup, conn *nats.EncodedConn) {
	syncSub, err := conn.Conn.SubscribeSync(MY_TOPIC)
	if err != nil {
		panic(err)
	}
	wg.Done()
	for i := 0; i < 30; i++ {
		getMsg, err := syncSub.NextMsg(time.Second)
		if err != nil {
			fmt.Println("同步消息订阅，nextMsg get err : " + err.Error())
			return
		}
		fmt.Println("同步拿消息成功" + string(getMsg.Data))
	}

}

func subscribeRequestWithInboxDemo(wg *sync.WaitGroup, inboxString string, conn *nats.EncodedConn) {
	sb, err := conn.Conn.SubscribeSync(MY_TOPIC_REQUEST)
	if err != nil {
		panic(err)
	}
	wg.Done()
	for i := 0; i < 30; i++ {
		getMsg, err := sb.NextMsg(time.Second * 5)
		if err != nil {
			fmt.Println("同步消息订阅，nextMsg get err : " + err.Error())
			return
		}
		getMsg.Respond([]byte("接受成功，发一个respond"))
	}
}

func sendToSyncRequestDemo(wg *sync.WaitGroup, inboxString string, conn *nats.EncodedConn) {
	wg.Wait()

	go func() {

		for i := 0; i < 10; i++ {
			err := conn.PublishRequest(MY_TOPIC_REQUEST, inboxString, "我是一条request消息")
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Second)
		}
	}()

	time.Sleep(time.Second * 2)

	// 订阅inbox，从inbox里拿消息
	sb, err := conn.Conn.SubscribeSync(inboxString)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 30; i++ {
		msg, err := sb.NextMsg(time.Second * 5)
		if err != nil {
			fmt.Println("从inbox拿消息捕捉错误 " + err.Error())
			return
		}
		fmt.Printf("%d 收到reply: %s \n", time.Now().Unix(), msg.Data)
	}
}
