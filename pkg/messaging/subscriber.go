package messaging

import (
	"fmt"
	"github.com/streadway/amqp"
	"minik8s/pkg/klog"
	"sync"
	"time"
)

type HandleFunc func(d amqp.Delivery)

type redo struct {
	exchangeName string
	handler      HandleFunc
	stopCh       <-chan struct{}
}

type Subscriber struct {
	conn          *amqp.Connection // conn RabbitMQ connection
	connUrl       string           // connUrl connection url
	maxRetry      int              // maxRetry 当连接意外中断时的最大重连次数
	retryInterval time.Duration    // retryInterval
	redoLogs      map[int]redo     // redoLogs 存储了redo log，在发生断线重连时需要进行redo，记录了subscribe的信息
	normal        bool             // normal 当手动关闭连接时为true，否则为false
	nextSlot      int              // nextSlot 下一个slot的id
	mtxRecover    sync.Mutex       // mtxRecover 恢复状态的锁
	mtxNormal     sync.Mutex       // mtxNormal 保护normal元数据
	mtxRedo       sync.Mutex       // mtxRedo 保护redoLogs元数据
}

/*
NewSubscriber 创建一个Subscriber并且返回指针
*/
func NewSubscriber(config QConfig) (*Subscriber, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.User, config.Password, config.Host, config.Port)
	var err error
	s := new(Subscriber)
	s.connUrl = url
	s.maxRetry = config.MaxRetry
	s.retryInterval = config.RetryInterval
	s.redoLogs = make(map[int]redo)
	s.normal = false
	s.nextSlot = 0
	s.conn, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	errCh := make(chan *amqp.Error)
	go s.rerun(errCh)
	s.conn.NotifyClose(errCh)
	return s, nil
}

/*
CloseConnection 关闭subscriber的connection，subscriber变为不可用状态
*/
func (s *Subscriber) CloseConnection() error {
	s.mtxNormal.Lock()
	s.normal = true
	s.mtxNormal.Unlock()
	if !s.conn.IsClosed() {
		err := s.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Subscribe 创建一个rabbitmq channel并且放入goroutine消费消息

goroutine直到stopCh收到消息才退出，使用Close(stopCh)来关闭rabbitmq channel

exchangeName 订阅的交换机名称

handler 传入的处理函数

*/
func (s *Subscriber) Subscribe(exchangeName string, handler HandleFunc, stopChannelCh <-chan struct{}) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return err
	}

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

	// 加入redo log中，便于后续断线重连的恢复
	s.mtxRedo.Lock()
	index := s.nextSlot
	s.nextSlot++
	s.redoLogs[index] = redo{exchangeName: exchangeName, handler: handler, stopCh: stopChannelCh}
	s.mtxRedo.Unlock()

	stop := func(amqpChannel *amqp.Channel, index int, stopChannelCh <-chan struct{}, stopConnectionCh <-chan *amqp.Error) {
		select {
		case <-stopConnectionCh:
			klog.Infof("connection closed!\n")
			return
		case <-stopChannelCh:
			s.mtxRedo.Lock()
			klog.Infof("remove redo log %d \n", index)
			delete(s.redoLogs, index)
			s.mtxRedo.Unlock()
			_ = amqpChannel.Close()
			return
		}
	}

	consumeLoop := func(deliveries <-chan amqp.Delivery, handler HandleFunc) {
		for d := range deliveries {
			// Invoke the handlerFunc func we passed as parameter.
			handler(d)
		}
	}

	stopConnectionCh := make(chan *amqp.Error)
	s.conn.NotifyClose(stopConnectionCh)
	go stop(ch, index, stopChannelCh, stopConnectionCh)
	go consumeLoop(msgs, handler)
	return nil
}

/*
Unsubscribe  取消一个之前已经订阅的subscribe

*/
func (s *Subscriber) Unsubscribe(stopChannelCh chan<- struct{}) {
	s.mtxRecover.Lock()
	defer s.mtxRecover.Unlock()
	close(stopChannelCh)
}

func (s *Subscriber) recover() {
	s.mtxRecover.Lock()
	defer s.mtxRecover.Unlock()
	s.mtxRedo.Lock()
	redoCopy := s.redoLogs
	s.redoLogs = make(map[int]redo)
	s.nextSlot = 0
	s.mtxRedo.Unlock()
	for _, redo := range redoCopy {
		err := s.Subscribe(redo.exchangeName, redo.handler, redo.stopCh)
		if err != nil {
			klog.Errorf("Error subscribing while reconnecting!\n")
		}
	}
}

func (s *Subscriber) reconnect() error {
	var err error
	s.conn, err = amqp.Dial(s.connUrl)
	if err != nil {
		return err
	}
	errCh := make(chan *amqp.Error)
	go s.rerun(errCh)
	s.conn.NotifyClose(errCh)
	s.recover()
	return nil
}

func (s *Subscriber) rerun(errCh <-chan *amqp.Error) {
	time.Sleep(time.Second)
	select {
	case <-errCh:
		s.mtxNormal.Lock()
		normal := s.normal
		s.mtxNormal.Unlock()
		if normal {
			klog.Infof("Subscriber : Close connection normally...\n")
			return
		}
		if !normal {
			for i := 1; i <= s.maxRetry; i++ {
				klog.Warnf("Subscriber : Trying to reconnect : retry - %d\n", i)
				if err := s.reconnect(); err == nil {
					klog.Infof("Subscriber reconnected!")
					return
				}
				time.Sleep(s.retryInterval)
			}
			// 在最大重试次数之后仍然无法重连
			klog.Errorf("Subscriber : Error reconnecting!\n")
		}
	}
}
