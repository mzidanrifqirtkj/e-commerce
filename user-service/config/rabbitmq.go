package config

import (
	"fmt"

	"github.com/streadway/amqp"
)

func (cfg Config) NewRabbitMQ() (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMQ.User, cfg.RabbitMQ.Port, cfg.RabbitMQ.Host, cfg.RabbitMQ.Password)
	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Errorf("[NewRabbitMQ-1] Failed to connect RabbitMQ: %v", err)
		return nil, err
	}

	return conn, nil
}
