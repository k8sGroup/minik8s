package messaging

import (
	"fmt"
	"github.com/streadway/amqp"
)

type Publisher struct {
	conn *amqp.Connection
}

/*
NewPublisher 创建一个Publisher并且返回其指针
*/
func NewPublisher(user string, password string, host string, port string) (*Publisher, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	publisher := new(Publisher)
	var err error
	publisher.conn, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return publisher, nil
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
