package locate

type Locator interface {
	Name() string
	// 绑定用户到网关节点
	BindGateNode(uid int64, nodeId string, connId int64) error
	// 解绑用户到网关节点
	UnbindGateNode(uid int64, nodeId string, connId int64) error
	// 获取用户所在的网关节点
	GetGateNode(uid int64) (string, error)
	// 获取用户所在的网关连接ID
	GetGateConnId(uid int64) (int64, error)
}
