package transceiver

type Transceiver interface {
	SendMessage(uid int64, payload []byte, gameName, node string) (string, error)
	ReceiveMessage(gameName, node string, handler func(uid int64, payload []byte, msgId string) error) error
	Close() error
}
