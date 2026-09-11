package log

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DefaultTailLines = 500
	MaxTailLines     = 5000
	rotateSize       = 5 * 1024 * 1024

	LogFileName      = "BSManager.log"
	CrashLogFileName = "crash.log"
)

var (
	logger *zap.Logger
	level  = zap.NewAtomicLevelAt(zap.InfoLevel)
	once   sync.Once

	mu           sync.RWMutex
	logFilePath  string
	crashLogPath string
)

func encoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg
}

func newLogger(core zapcore.Core) *zap.Logger {
	return zap.New(core, zap.AddStacktrace(zap.FatalLevel))
}

func init() {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig()),
		zapcore.Lock(os.Stderr),
		level,
	)
	logger = newLogger(core)
}

func parseLevel(l string) zapcore.Level {
	switch l {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func SetLogLevel(l string) {
	level.SetLevel(parseLevel(l))
}

func SetLogFile(path string) error {
	mu.Lock()
	defer mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if st, err := os.Stat(path); err == nil && st.Size() > rotateSize {
		_ = os.Rename(path, path+".old")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig()),
		zapcore.Lock(f),
		level,
	)
	logger = newLogger(zapcore.NewTee(logger.Core(), fileCore))
	logFilePath = path
	AddCloseFunc(f.Close)
	return nil
}

func SetCrashLog(path string) error {
	mu.Lock()
	defer mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if st, err := os.Stat(path); err == nil && st.Size() > rotateSize {
		_ = os.Rename(path, path+".old")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := debug.SetCrashOutput(f, debug.CrashOptions{}); err != nil {
		return err
	}
	crashLogPath = path
	return nil
}

func CrashLogPath() string {
	mu.RLock()
	defer mu.RUnlock()
	return crashLogPath
}

func LogFilePath() string {
	mu.RLock()
	defer mu.RUnlock()
	return logFilePath
}

func ReadTail(lines int) ([]string, error) {
	path := LogFilePath()
	if path == "" {
		return nil, errors.New("log: log file is not set")
	}
	if lines <= 0 {
		lines = DefaultTailLines
	}
	if lines > MaxTailLines {
		lines = MaxTailLines
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ring := make([]string, lines)
	count := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		ring[count%lines] = sc.Text()
		count++
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	n := min(count, lines)
	out := make([]string, 0, n)
	start := count - n
	for i := start; i < count; i++ {
		out = append(out, ring[i%lines])
	}
	return out, nil
}

func GetLogger() *zap.Logger {
	return logger
}

var CloseFunc []func() error

func AddCloseFunc(f func() error) {
	CloseFunc = append(CloseFunc, f)
}

func Close() {
	once.Do(func() {
		_ = logger.Sync()
		for _, f := range CloseFunc {
			err := f()
			if err != nil {
				logger.Error("close error", zap.Error(err))
			}
		}
	})
}

func SubLogger(name string) *zap.Logger {
	return logger.Named(name)
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func ErrorE(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
	Close()
	os.Exit(1)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}
