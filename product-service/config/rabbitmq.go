package config

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func (cfg Config) NewRabbitMQ() (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMQ.User, cfg.RabbitMQ.Password, cfg.RabbitMQ.Host, cfg.RabbitMQ.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("[NewRabbitMQ-1] Failed to connect RabbitMQ: %v", err)
		return nil, err
	}

	return conn, nil
}
