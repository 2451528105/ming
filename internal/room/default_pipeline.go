package room

import "ming/internal/service"

// NewDefaultPipeline 使用默认解析器与房间业务服务构造一条 Pipeline（需调用方持有；进程级单例见 InitDefault）。
func NewDefaultPipeline() (*Pipeline, error) {
	return Init(PipelineOptions{
		MailboxSize: 256,
		Resolver:    NewUIDResolver(),
		Service:     service.NewDefaultRoomService(),
	})
}
