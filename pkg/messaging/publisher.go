package messaging

import (
	"fmt"
	"github.com/streadway/amqp"
	"minik8s/pkg/klog"
	"sync"
	"time"
)

type Publisher struct {
	conn          *amqp.Connection
	connUrl       string
	maxRetry      int
	retryInterval time.Duration
	normal        bool
	mtxNormal     sync.Mutex
}

/*
NewPublisher 创建一个Publisher并且返回其指针
*/
func NewPublisher(config *QConfig) (*Publisher, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.User, config.Password, config.Host, config.Port)
	p := new(Publisher)
	var err error
	p.connUrl = url
	p.maxRetry = config.MaxRetry
	p.retryInterval = config.RetryInterval
	p.normal = false
	p.conn, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	errCh := make(chan *amqp.Error)
	go p.rerun(errCh)
	p.conn.NotifyClose(errCh)
	return p, nil
}

/*
Publish 向指定的交换机广播一条信息并立即返回，广播类型为FANOUT

*/
func (p *Publisher) Publish(exchangeName string, body []byte, contentType string) error {

	ch, err := p.conn.Channel()
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

	err = ch.Publish(
		exchangeName,
		exchangeName,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        body,
		})
	if err != nil {
		return err
	}
	return nil
}

func (p *Publisher) CloseConnection() error {
	p.mtxNormal.Lock()
	p.normal = true
	p.mtxNormal.Unlock()
	if !p.conn.IsClosed() {
		err := p.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Publisher) rerun(errCh <-chan *amqp.Error) {
	time.Sleep(time.Second)
	select {
	case <-errCh:
		p.mtxNormal.Lock()
		normal := p.normal
		p.mtxNormal.Unlock()
		if normal {
			klog.Infof("Publisher : Close connection normally...\n")
			return
		}
		if !normal {
			for i := 1; i <= p.maxRetry; i++ {
				klog.Warnf("Publisher : Trying to reconnect : retry - %d\n", i)
				if err := p.reconnect(); err == nil {
					klog.Infof("Publisher : reconnected!")
					return
				}
				time.Sleep(p.retryInterval)
			}
			// 在最大重试次数之后仍然无法重连
			klog.Errorf("Publisher : Error reconnecting!\n")
		}
	}
}

func (p *Publisher) reconnect() error {
	var err error
	p.conn, err = amqp.Dial(p.connUrl)
	if err != nil {
		return err
	}
	errCh := make(chan *amqp.Error)
	go p.rerun(errCh)
	p.conn.NotifyClose(errCh)
	return nil
}
