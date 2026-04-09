package config

import "time"

type Config struct {
	Application ApplicationConfig `yaml:"Application"` // 应用配置
	Log         LoggerConfig      `yaml:"Log"`         // 日志配置
	Redis       RedisConfig       `yaml:"Redis"`       // redis配置
	PostgresSql PostgresSqlConfig `yaml:"PostgresSql"` // postgres配置
	RocketMQ    RocketMQConfig    `yaml:"RocketMQ"`    // rocketmq配置
}

var Cfg *Config = &Config{}

type LoggerConfig struct {
	MinLevel   string        `yaml:"MinLevel"`   // 日志输出最低级别
	Pathname   string        `yaml:"Pathname"`   // 日志文件路径
	Interval   time.Duration `yaml:"Interval"`   // 日志文件切割时间间隔
	TimeFormat string        `yaml:"TimeFormat"` // 时间格式
}

type ApplicationConfig struct {
	Name          string `yaml:"Name"`          // 服务名
	GameCode      string `yaml:"GameCode"`      // 游戏编码
	Env           string `yaml:"Env"`           // 环境
	WebsocketPath string `yaml:"WebsocketPath"` // websocket路径
	Port          int    `yaml:"Port"`          // 端口
}

type RedisConfig struct {
	Addr       string `yaml:"Addr"`       //redis地址
	ClientName string `yaml:"ClientName"` //redis客户端名称
	Username   string `yaml:"Username"`   //redis用户名
	Password   string `yaml:"Password"`   //redis密码
	DB         int    `yaml:"DB"`         //redis数据库
}

type PostgresSqlConfig struct {
	Host     string `yaml:"Host"`     // postgres主机
	Port     int    `yaml:"Port"`     // postgres端口
	User     string `yaml:"User"`     // postgres用户名
	Password string `yaml:"Password"` // postgres密码
	DB       string `yaml:"DB"`       // postgres数据库
}

type RocketMQConfig struct {
	Endpoint      string `yaml:"Endpoint"`      // rocketmq地址
	Namespace     string `yaml:"Namespace"`     // rocketmq命名空间
	ConsumerGroup string `yaml:"ConsumerGroup"` // rocketmq消费者组
}
