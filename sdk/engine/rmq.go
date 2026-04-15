package engine

import (
	"fmt"
	"ming/sdk/xlog"
	"time"

	rmq "github.com/apache/rocketmq-clients/golang"
	"github.com/apache/rocketmq-clients/golang/credentials"
)

func InitRmq(endpoint, namespace, group string) (rmq.Producer, rmq.SimpleConsumer, error) {
	producer, err := NewRMQProducer(endpoint, namespace, group)
	if err != nil {
		xlog.Error().Err(err).Msg("初始化RMQ生产者失败")
		return nil, nil, err
	}
	consumer, err := NewRMQConsumer(endpoint, namespace, group)
	if err != nil {
		xlog.Error().Err(err).Msg("初始化RMQ消费者失败")
		_ = producer.GracefulStop()
		return nil, nil, err
	}
	// 统一在初始化阶段启动一次，后续业务层直接使用，避免重复 Start。
	if err = consumer.Start(); err != nil {
		xlog.Error().Err(err).Msg("启动RMQ消费者失败")
		_ = consumer.GracefulStop()
		_ = producer.GracefulStop()
		return nil, nil, err
	}

	return producer, consumer, nil
}

// NewRMQProducer 新建RMQ生产者
func NewRMQProducer(endpoint, namespace, group string) (rmq.Producer, error) {

	p, err := rmq.NewProducer(&rmq.Config{
		Endpoint:      endpoint,
		NameSpace:     namespace,
		ConsumerGroup: group,
		Credentials:   &credentials.SessionCredentials{},
	})
	if err != nil {
		return nil, fmt.Errorf("rmq.NewProducer err: %v", err)
	}
	return p, nil
}

// NewRMQConsumer 新建RMQ消费者
func NewRMQConsumer(endpoint, namespace, group string) (rmq.SimpleConsumer, error) {

	c, err := rmq.NewSimpleConsumer(
		&rmq.Config{
			Endpoint:      endpoint,
			NameSpace:     namespace,
			ConsumerGroup: group,
			Credentials:   &credentials.SessionCredentials{},
		},
		rmq.WithAwaitDuration(time.Second*5),
	)
	if err != nil {
		return nil, fmt.Errorf("rmq.NewSimpleConsumer err: %v", err)
	}
	return c, nil
}
