package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Logger 全局日志实例
	Logger *zap.Logger
	// SugarLogger 语法糖日志实例
	SugarLogger *zap.SugaredLogger
)

// Setup 初始化日志
func Setup() error {
	// 从配置中获取日志设置
	logLevel := viper.GetString("log.level")
	logFormat := viper.GetString("log.format")
	logOutput := viper.GetString("log.output")
	logFilePath := viper.GetString("log.file_path")
	logMaxSize := viper.GetInt("log.max_size")
	logMaxAge := viper.GetInt("log.max_age")
	logMaxBackups := viper.GetInt("log.max_backups")
	logCompress := viper.GetBool("log.compress")

	// 设置日志级别
	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 设置日志编码器
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if logFormat == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置日志输出
	var writeSyncer zapcore.WriteSyncer
	if logOutput == "file" {
		// 确保日志目录存在
		logDir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// 使用lumberjack进行日志轮转
		lumberJackLogger := &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    logMaxSize,    // 单位：MB
			MaxBackups: logMaxBackups, // 最大备份数
			MaxAge:     logMaxAge,     // 最大保存天数
			Compress:   logCompress,   // 是否压缩
		}
		writeSyncer = zapcore.AddSync(lumberJackLogger)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	SugarLogger = Logger.Sugar()

	return nil
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Debugf 调试日志（格式化）
func Debugf(format string, args ...interface{}) {
	SugarLogger.Debugf(format, args...)
}

// Infof 信息日志（格式化）
func Infof(format string, args ...interface{}) {
	SugarLogger.Infof(format, args...)
}

// Warnf 警告日志（格式化）
func Warnf(format string, args ...interface{}) {
	SugarLogger.Warnf(format, args...)
}

// Errorf 错误日志（格式化）
func Errorf(format string, args ...interface{}) {
	SugarLogger.Errorf(format, args...)
}

// Fatalf 致命错误日志（格式化）
func Fatalf(format string, args ...interface{}) {
	SugarLogger.Fatalf(format, args...)
}

// Sync 同步日志，应用退出前调用
func Sync() {
	_ = Logger.Sync()
	_ = SugarLogger.Sync()
}