package plugin

import (
	"fmt"
	"strings"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
	"github.com/spf13/cast"
	"github.com/twmb/franz-go/pkg/kgo"
)

type kgoLogAdapter struct {
	level kgo.LogLevel
}

func NewKgoLogAdapter(level kgo.LogLevel) kgo.Logger {
	return &kgoLogAdapter{level: level}
}

func (l *kgoLogAdapter) Level() kgo.LogLevel {
	return l.level
}

func (l *kgoLogAdapter) Log(level kgo.LogLevel, msg string, keyvals ...any) {
	handlerLevel := l.convKgoLevelToHostLevel(level)
	if len(keyvals)%2 != 0 {
		handler.Host.Log(handlerLevel, fmt.Sprintf("msg:%q", msg))
		return
	}

	var kvs strings.Builder
	for i := 0; i < len(keyvals); i += 2 {
		kvs.WriteString("\"" + cast.ToString(keyvals[i]) + "\"")
		kvs.WriteRune(':')
		kvs.WriteString("\"" + cast.ToString(keyvals[i+1]) + "\"")

		if i < len(keyvals)-2 {
			kvs.WriteRune(',')
		}
	}

	handler.Host.Log(l.convKgoLevelToHostLevel(level), fmt.Sprintf("msg:%q,%s", msg, kvs.String()))
}

func (l *kgoLogAdapter) convKgoLevelToHostLevel(level kgo.LogLevel) api.LogLevel {
	switch level {
	case kgo.LogLevelNone:
		return api.LogLevelNone
	case kgo.LogLevelError:
		return api.LogLevelError
	case kgo.LogLevelWarn:
		return api.LogLevelWarn
	case kgo.LogLevelDebug:
		return api.LogLevelDebug
	default:
		return api.LogLevelInfo
	}
}
