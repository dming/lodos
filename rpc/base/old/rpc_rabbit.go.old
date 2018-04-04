package baserpc

import (
	"fmt"
	"github.com/dming/lodos/conf"
	log "github.com/dming/lodos/log"
	"github.com/streadway/amqp"
)

type Consumer struct {
	info           *conf.Rabbitmq
	conn           *amqp.Connection
	channel        *amqp.Channel
	callback_queue string
	tag            string
}

type RabbitMQInfo struct {
	uri            string
	exchange       string
	exchangeType   string
	queue          string
	callback_queue string
	bindingKey     string //
	consumerTag    string //消费者TAG
}

func NewConsumer(info *conf.Rabbitmq, amqpURI, exchange, exchangeType, ctag string) (*Consumer, error) {
	c := &Consumer{
		info:           info,
		conn:           nil,
		channel:        nil,
		callback_queue: "",
		tag:            ctag,
	}

	var err error

	log.Info("dialing %q", amqpURI)
	c.conn, err = amqp.Dial(amqpURI) //打开连接
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	//go func() {
	//	fmt.Printf("amqp closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	//}()

	log.Info("got Connection, getting Channel")
	//声明channel
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Info("got Channel, declaring Exchange (%q)", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	return c, nil
}

func (c *Consumer) Cancel() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}
	return nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	return nil
}

