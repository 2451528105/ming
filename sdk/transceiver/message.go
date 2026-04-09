package transceiver

type message struct {
	UUID      string `json:"uuid"`      //消息唯一ID
	Uid       int64  `json:"uid"`       //用户ID
	Timestamp int64  `json:"timestamp"` //消息时间戳
	Payload   []byte `json:"payload"`   //消息内容
}
