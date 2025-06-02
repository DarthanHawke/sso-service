package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Возвращает new Logger из настраиваемого zapcore.Core
func SetupLogger() *zap.Logger {
	// синхронная запиись в Stdout
	stdout := zapcore.AddSync(os.Stdout)
	// синхронная запиись в файл с использованием lumberjack для ротации файлов журнала
	logFile := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
	})

	prodConfig := zap.NewProductionEncoderConfig()
	prodConfig.TimeKey = "timestamp"
	prodConfig.EncodeTime = zapcore.ISO8601TimeEncoder // меняем ts поле на timestamp и формат времени на ISO-8601

	devConfig := zap.NewDevelopmentEncoderConfig()
	devConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // цветной вывод логов в консоль

	consoleEncoder := zapcore.NewConsoleEncoder(devConfig)
	fileEncoder := zapcore.NewJSONEncoder(prodConfig)

	// Дублирует записи журнала в консоль и в файл
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, zapcore.InfoLevel),
		zapcore.NewCore(fileEncoder, logFile, zapcore.DebugLevel),
	)

	// zap.AddCaller() - добавляем к каждому сообщению имя файла, номер строки и название функции вызывающей zap
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
