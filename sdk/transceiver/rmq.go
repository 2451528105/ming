package transceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	rmq "github.com/apache/rocketmq-clients/golang"
	"github.com/google/uuid"
)

type XRMQTransceiver struct {
	topic    string
	producer rmq.Producer
	consumer rmq.SimpleConsumer
	ctx      context.Context
	cancel   context.CancelFunc
	once     sync.Once
	wg       sync.WaitGroup
	subs     sync.Map // key: topic, value: func(msg *rmq.MessageView) error
}

// safeSubscribe wraps third-party SDK panic as error to avoid process crash.
func safeSubscribe(consumer rmq.SimpleConsumer, topic, tag string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("rmq subscribe panic, topic=%s, tag=%s, panic=%v, stack=%s", topic, tag, r, string(debug.Stack()))
		}
	}()
	return consumer.Subscribe(topic, rmq.NewFilterExpression(tag))
}

// 新建XRMQTransceiver
func NewXRMQTransceiver(topic string, producer rmq.Producer, consumer rmq.SimpleConsumer) *XRMQTransceiver {
	ctx, cancel := context.WithCancel(context.Background())
	return &XRMQTransceiver{
		topic:    topic,
		producer: producer,
		consumer: consumer,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// 获取消息标签
func getTag(gameName, node string) string {
	return fmt.Sprintf("%s-%s", gameName, node)
}

// 发送消息
func (t *XRMQTransceiver) SendMessage(uid int64, payload []byte, gameName, node string) (string, error) {
	tag := getTag(gameName, node)
	m := &message{
		UUID:      uuid.New().String(),
		Uid:       uid,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	}
	body, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("json marshal message: %w", err)
	}
	msg := &rmq.Message{
		Topic: t.topic,
		Tag:   &tag,
		Body:  body,
	}
	receipt, err := t.producer.Send(context.Background(), msg)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}
	return receipt[0].MessageID, nil
}

// 接收消息
func (t *XRMQTransceiver) ReceiveMessage(gameName, node string, handler func(uid int64, payload []byte, msgId string) error) error {
	tag := getTag(gameName, node)
	cb := func(msg *rmq.MessageView) error {
		var body message
		if err := json.Unmarshal(msg.GetBody(), &body); err != nil {
			return fmt.Errorf("json unmarshal message: %w", err)
		}
		return handler(body.Uid, body.Payload, body.UUID)
	}
	if _, ok := t.subs.LoadOrStore(t.topic, cb); ok {
		return fmt.Errorf("topic %s has been subscribed", t.topic)
	}

	if err := safeSubscribe(t.consumer, t.topic, tag); err != nil {
		t.subs.Delete(t.topic)
		return fmt.Errorf("subscribe topic %s failed: %w", t.topic, err)
	}
	t.once.Do(func() {
		t.wg.Add(1)
		go t.watch()
	})

	return nil
}

func (t *XRMQTransceiver) watch() {
	defer t.wg.Done()
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}
		msgs, err := t.consumer.Receive(t.ctx, 16, 30*time.Second)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		for _, msg := range msgs {
			cbRaw, ok := t.subs.Load(msg.GetTopic())
			if !ok || cbRaw == nil {
				continue
			}
			cb, ok := cbRaw.(func(msg *rmq.MessageView) error)
			if !ok {
				continue
			}
			if err = cb(msg); err != nil {
				continue
			}
			_ = t.consumer.Ack(t.ctx, msg)
		}
	}
}

func (t *XRMQTransceiver) Close() error {
	if t.producer != nil {
		if err := t.producer.GracefulStop(); err != nil {
			return err
		}
	}
	if t.consumer != nil {
		if err := t.consumer.GracefulStop(); err != nil {
			return err
		}
	}
	return nil
}
