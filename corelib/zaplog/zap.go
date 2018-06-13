package zaplog

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = (*ZapSugar)(nil)

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("01/02 15:04:05.999"))
}

// ZapSugar implements Logger with zap sugar logger
type ZapSugar struct {
	prefix string
	sugar  *zap.SugaredLogger
}

func (zs *ZapSugar) initSugar(isDebug bool) {
	config := zap.NewProductionConfig()
	lv := zap.NewAtomicLevel()
	lv.SetLevel(zap.InfoLevel)
	if isDebug {
		lv.SetLevel(zap.DebugLevel)
		// config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	config.Sampling = &zap.SamplingConfig{
		Initial:    1e5,
		Thereafter: 1e5,
	}
	config.Level = lv
	config.Encoding = "console"
	config.EncoderConfig.EncodeTime = timeEncoder
	config.DisableCaller = true
	// config.DisableStacktrace = true

	l, err := config.Build()
	if err != nil {
		panic("log init error : " + err.Error())
	}
	if zs.sugar != nil {
		su := l.Sugar()
		*zs.sugar = *su
		return
	}
	zs.sugar = l.Sugar()
}

// NewZapSugar return logger with prefix string, and debug level setting
// if prefix is set, print log message like 'prefix : message'
// if set debug to true, print debug level logs. otherwise, print info level and above
func NewZapSugar(prefix string, debug bool) Logger {
	z := &ZapSugar{
		prefix: prefix,
	}
	z.initSugar(debug)
	return z
}

// Sync sync logs
func (zs *ZapSugar) Sync() {
	zs.sugar.Sync()
}

func (zs *ZapSugar) prefixMsg(msg string) string {
	if zs.prefix == "" {
		return msg
	}
	return zs.prefix + " : " + msg
}

// Debug logs message as debug level
func (zs *ZapSugar) Debug(msg string, values ...interface{}) {
	msg = zs.prefixMsg(msg)
	if len(values) == 0 {
		zs.sugar.Debug(msg)
		return
	}
	zs.sugar.Debugw(msg, values...)
}

// Info logs message as info level
func (zs *ZapSugar) Info(msg string, values ...interface{}) {
	msg = zs.prefixMsg(msg)
	if len(values) == 0 {
		zs.sugar.Info(msg)
		return
	}
	zs.sugar.Infow(msg, values...)
}

// Warn logs message as warn level
func (zs *ZapSugar) Warn(msg string, values ...interface{}) {
	msg = zs.prefixMsg(msg)
	if len(values) == 0 {
		zs.sugar.Warn(msg)
		return
	}
	zs.sugar.Warnw(msg, values...)
}

// Error logs message as error level
// it logs caller stack together
func (zs *ZapSugar) Error(msg string, values ...interface{}) {
	msg = zs.prefixMsg(msg)
	if len(values) == 0 {
		zs.sugar.Error(msg)
		return
	}
	zs.sugar.Errorw(msg, values...)
}

// Fatal logs message as fatal level
// it calls os.Exit(1) after logged
func (zs *ZapSugar) Fatal(msg string, values ...interface{}) {
	msg = zs.prefixMsg(msg)
	if len(values) == 0 {
		zs.sugar.Fatal(msg)
		return
	}
	zs.sugar.Fatalw(msg, values...)
}

// NewPrefix return new suger logger with prefix message
func (zs *ZapSugar) NewPrefix(prefix string) Logger {
	return &ZapSugar{
		sugar:  zs.sugar,
		prefix: prefix,
	}
}

// SetDebug sets logger to print debug
func (zs *ZapSugar) SetDebug(debug bool) {
	zs.initSugar(debug)
}

var (
	gZap = NewZapSugar("log", false)
)

// Zap returns global zap logger with new prefix
func Zap(prefix string) Logger {
	return gZap.NewPrefix(prefix)
}
