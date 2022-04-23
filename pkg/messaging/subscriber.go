package messaging

import (
	"fmt"
	"github.com/streadway/amqp"
	"minik8s/pkg/klog"
	"sync"
	"time"
)

const rerunDuration = 10 * time.Second

type HandleFunc func(d amqp.Delivery)

type SubscriberConfig struct {
	user     string
	password string
	host     string
	port     string
	maxRetry int
}

type redo struct {
	exchangeName string
	handler      HandleFunc
	stopCh       <-chan struct{}
}

type Subscriber struct {
	// conn RabbitMQ connection
	conn *amqp.Connection
	// connUrl connection url
	connUrl string
	// maxRetry 当连接意外中断时的最大重连次数
	maxRetry int
	// errCh error channel用于发生意外close时的notify rerun机制
	errCh chan *amqp.Error
	// redoLogs 存储了redo log，在发生断线重连时需要进行redo，记录了subscribe的信息
	redoLogs map[int]redo
	// normal 当手动关闭连接时为true，否则为false
	normal bool
	// nextSlot 下一个slot的id
	nextSlot int
	// mtxNormal 保护normal元数据
	mtxNormal sync.Mutex
	// mtxRedo 保护redoLogs元数据
	mtxRedo sync.Mutex
}

/*
NewSubscriber 创建一个Subscriber并且返回指针

*/
func NewSubscriber(config SubscriberConfig) (*Subscriber, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.user, config.password, config.host, config.port)
	var err error
	s := new(Subscriber)
	s.connUrl = url
	s.maxRetry = config.maxRetry
	s.errCh = make(chan *amqp.Error)
	s.redoLogs = make(map[int]redo)
	s.normal = false
	s.nextSlot = 0
	s.conn, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	reconnect := func() error {
		var err error
		s.conn, err = amqp.Dial(s.connUrl)
		if err != nil {
			return err
		}
		s.conn.NotifyClose(s.errCh)
		return nil
	}

	rerun := func(errCh <-chan *amqp.Error) {
		time.Sleep(5 * rerunDuration)
		for {
			select {
			case <-errCh:
				s.mtxNormal.Lock()
				normal := s.normal
				s.mtxNormal.Unlock()
				if !s.conn.IsClosed() {
					continue
				}
				if normal {
					klog.Infof("Close connection normally...\n")
					return
				}
				if !normal {
					for i := 1; i <= s.maxRetry; i++ {
						klog.Infof("Trying to reconnect : retry - %d\n", i)
						if err := reconnect(); err == nil {
							s.redoAll()
							s.conn.NotifyClose(s.errCh)
							break
						}
						time.Sleep(rerunDuration)
					}
					// 在最大重试次数之后仍然无法重连
					klog.Fatalf("Error reconnecting!\n")
				}
			}
		}
	}

	go rerun(s.errCh)
	s.conn.NotifyClose(s.errCh)
	return s, nil
}

/*
CloseConnection 关闭Subscriber的connection，subscriber变为不可用状态

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
Subscribe 创建一个rabbitmq channel并且放入goroutine消费消息，
goroutine直到stopCh收到消息才退出，使用Close(stopCh)来关闭rabbitmq channel

exchangeName 订阅的交换机名称

handler 传入的处理函数

*/
func (s *Subscriber) Subscribe(exchangeName string, handler HandleFunc, stopCh <-chan struct{}) error {
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
	s.redoLogs[index] = redo{exchangeName: exchangeName, handler: handler, stopCh: stopCh}
	s.mtxRedo.Unlock()

	stop := func(ch *amqp.Channel, stopCh <-chan struct{}, index int, stopConnCh <-chan *amqp.Error) {
		select {
		case <-stopConnCh:
			return
		case <-stopCh:
			s.mtxRedo.Lock()
			delete(s.redoLogs, index)
			s.mtxRedo.Unlock()
			_ = ch.Close()
			return
		}
	}

	consumeLoop := func(deliveries <-chan amqp.Delivery, handler HandleFunc) {
		for d := range deliveries {
			// Invoke the handlerFunc func we passed as parameter.
			handler(d)
		}
	}

	stopConnCh := make(chan *amqp.Error)
	s.conn.NotifyClose(stopConnCh)
	go stop(ch, stopCh, index, stopConnCh)
	go consumeLoop(msgs, handler)
	return nil
}

func (s *Subscriber) redoAll() {
	s.mtxRedo.Lock()
	redoCopy := s.redoLogs
	s.redoLogs = make(map[int]redo)
	s.nextSlot = 0
	s.mtxRedo.Unlock()
	for _, redo := range redoCopy {
		err := s.Subscribe(redo.exchangeName, redo.handler, redo.stopCh)
		if err != nil {
			klog.Errorf("Error subscribing while reconnection!\n")
		}
	}
}
