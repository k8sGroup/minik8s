package messaging

import (
	"fmt"
	"github.com/streadway/amqp"
)

type HandleFunc func(d amqp.Delivery)

type Subscriber struct {
	conn *amqp.Connection
}

/*
NewSubscriber 创建一个Subscriber并且返回其指针

*/
func NewSubscriber(user string, password string, host string, port string) (*Subscriber, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	subscriber := new(Subscriber)
	var err error
	subscriber.conn, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return subscriber, nil
}

/*
Subscribe 从不退出，直到stopCh收到消息

exchangeName 订阅的交换机名称

handler 传入的处理函数

*/
func (s *Subscriber) Subscribe(exchangeName string, handler HandleFunc, stopCh <-chan struct{}) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		return err
	}

	queue, err := ch.QueueDeclare(
		"",
		false,
		true,
		false,
		false,
		nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		queue.Name,
		exchangeName,
		exchangeName,
		false,
		nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return err
	}

	go consumeLoop(msgs, handler)
	<-stopCh
	return nil
}

func consumeLoop(deliveries <-chan amqp.Delivery, handler HandleFunc) {
	for d := range deliveries {
		// Invoke the handlerFunc func we passed as parameter.
		handler(d)
	}
}
