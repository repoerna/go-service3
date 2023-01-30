package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Create application logger
func New(service string) (*zap.SugaredLogger, error) {
	conf := zap.NewProductionConfig()
	conf.OutputPaths = []string{"stdout"}
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	conf.DisableStacktrace = true
	conf.InitialFields = map[string]interface{}{
		"service": service,
	}

	log, err := conf.Build()
	if err != nil {
		return nil, err
	}

	return log.Sugar(), nil
}
