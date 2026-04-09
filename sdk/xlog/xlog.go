package xlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var logger *XLogger

type XLogger struct {
	std          *log.Logger
	writer       io.WriteCloser
	mux          sync.Mutex
	level        Level
	timeFormat   string
	interval     time.Duration // 日志切割时间间隔, 单位:h
	lastFileTime time.Time     // 上次log文件创建时间
	path         string        // 日志文件存放路径
	env          string        // 环境
	serviceName  string        // 服务名
	node         string        // 节点
	ip           string        // ip
}

type Level int

const (
	NoLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

func parseLevel(level string) Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	case "panic":
		return PanicLevel
	default:
		return DebugLevel
	}
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	default:
		return "log"
	}
}

type Event struct {
	logger   *XLogger
	level    Level
	fields   map[string]any
	ts       bool
	captured error
}

/*
初始化日志
level: 日志输出最低级别
pathname: 日志文件路径, 当为空时, 输出到控制台
interval: 日志切割时间间隔
serviceName: 服务名
env: 环境, 如: dev, test, prod
node: 节点, 如: node1, node2
ip: 节点ip, 如: 192.168.1.1
timeFormat: 时间格式, 为空时使用 RFC3339Nano
*/
func Init(level, pathname string, interval time.Duration, serviceName, env, node, ip, timeFormat string) {
	output := newOutput(pathname, node)
	if strings.TrimSpace(timeFormat) == "" {
		timeFormat = time.RFC3339Nano
	}

	logger = &XLogger{
		std:          log.New(output, "", 0),
		writer:       output,
		level:        parseLevel(level),
		timeFormat:   timeFormat,
		interval:     interval,
		mux:          sync.Mutex{},
		lastFileTime: time.Now(),
		path:         pathname,
		env:          env,
		serviceName:  serviceName,
		node:         node,
		ip:           ip,
	}
}

// 获取输出 控制台/文件
// 当pathname为空时,输出到控制台
func newOutput(pathname, node string) io.WriteCloser {

	// 1. 默认标准输出
	// 2. 文件夹设置不为空时,写入文件
	if pathname != "" {
		filename := "logs-node.log"
		if node != "" {
			filename = fmt.Sprintf("logs-%s.log", node)
		}
		// 文件夹不存在,则创建
		if _, err := os.Stat(pathname); os.IsNotExist(err) {
			err := os.MkdirAll(pathname, os.ModePerm)
			if err != nil {
				fmt.Println("MkdirAll path[", pathname, "] error:", err.Error())
			}
		}
		file, err := os.Create(path.Join(pathname, filename))
		if err == nil {
			return file
		} else {
			fmt.Println("create file[", filename, "] error:", err.Error())
		}
	}
	return nopWriteCloser{Writer: os.Stdout}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

func Debug() *Event {
	return newEvent(DebugLevel)
}

func Info() *Event {
	return newEvent(InfoLevel)
}

func Error() *Event {
	return newEvent(ErrorLevel)
}

func Warn() *Event {
	return newEvent(WarnLevel)
}

// Fatal Fatal消息打印 (程序终止)
func Fatal() *Event {
	return newEvent(FatalLevel)
}

// Panic Panic消息打印 (程序触发 panic)
func Panic() *Event {
	return newEvent(PanicLevel)
}

func logEvent() *Event {
	return newEvent(NoLevel)
}

func newEvent(level Level) *Event {
	// 1.检测logger引擎是否初始化、是否要切割
	logger.check()
	return (&Event{
		logger: logger,
		level:  level,
		fields: map[string]any{
			"env":     logger.env,
			"service": logger.serviceName,
			"ip":      logger.ip,
			"node":    logger.node,
		},
		ts: true,
	})
}

func (log *XLogger) check() {
	if log == nil {
		Init("debug", "", 0, "unknown", "unknown", "", "0.0.0.0", time.RFC3339Nano)
	} else {
		// 日志文件切割
		if log.interval > 0 && time.Now().Add(-log.interval).After(log.lastFileTime) {
			log.mux.Lock()
			if log.writer != nil {
				_ = log.writer.Close()
			}
			newWriter := newOutput(log.path, log.node)
			log.writer = newWriter
			log.std.SetOutput(newWriter)
			log.lastFileTime = time.Now()
			log.mux.Unlock()
		}
	}
}

func (e *Event) Str(key, value string) *Event {
	e.fields[key] = value
	return e
}

func (e *Event) Int(key string, value int) *Event {
	e.fields[key] = value
	return e
}

func (e *Event) Bool(key string, value bool) *Event {
	e.fields[key] = value
	return e
}

func (e *Event) Any(key string, value any) *Event {
	e.fields[key] = value
	return e
}

func (e *Event) Err(err error) *Event {
	if err != nil {
		e.captured = err
		e.fields["error"] = err.Error()
	}
	return e
}

func (e *Event) Timestamp() *Event {
	e.ts = true
	return e
}

func (e *Event) Msgf(format string, args ...any) {
	e.Msg(fmt.Sprintf(format, args...))
}

func (e *Event) Msg(msg string) {
	if e.logger == nil {
		return
	}
	if e.level != NoLevel && e.level < e.logger.level {
		return
	}
	if e.ts {
		e.fields["time"] = time.Now().Format(e.logger.timeFormat)
	}
	e.fields["level"] = e.level.String()
	e.fields["msg"] = msg

	keys := make([]string, 0, len(e.fields))
	for k := range e.fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(formatValue(e.fields[k]))
	}
	e.logger.std.Println(b.String())

	switch e.level {
	case FatalLevel:
		os.Exit(1)
	case PanicLevel:
		panic(msg)
	}
}

func formatValue(v any) string {
	switch t := v.(type) {
	case string:
		return strconv.Quote(t)
	case fmt.Stringer:
		return strconv.Quote(t.String())
	default:
		return fmt.Sprintf("%v", v)
	}
}
