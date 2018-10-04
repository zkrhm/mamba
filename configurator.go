package logconfig

import (
	"os"
	"path"
	"strings"

	"github.com/getsentry/raven-go"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	yaml "gopkg.in/yaml.v2"
)

//NewZapLogger returns configured zap logger.
func NewZapLogger(loggingConfig []byte) (*zap.Logger, error) {
	var config zap.Config

	if err := yaml.Unmarshal(loggingConfig, &config); err != nil {
		panic(err)
	}
	var paths []string
	//merge path of output and error output
	paths = append(paths, config.OutputPaths...)
	paths = append(paths, config.ErrorOutputPaths...)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := config.Build()

	for _, file := range paths {
		logger.Debug("creating paths ", zap.String("file", file))
		dir := path.Dir(file)
		// ignore if not file path (stdout / stderr)
		logger.Debug("dir is ", zap.String("dir ", dir))
		if dir != "." {
			//create dirs and file if not already existed.
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				logger.Debug("	making dir : Yes")
				if err := os.MkdirAll(dir, os.ModePerm); err != nil {
					logger.Error(err.Error())
					raven.CaptureError(err, nil)
				}
				if err := os.Mkdir(dir, os.ModePerm); err != nil {
					logger.Error(err.Error())
					raven.CaptureError(err, nil)
				}
			}

			if _, err := os.Stat(file); os.IsNotExist(err) {
				logger.Debug("	touching file : Yes")
				os.OpenFile(file, os.O_CREATE|os.O_APPEND, 0755)
			}
		}
	}

	// config.OutputPaths
	// config.ErrorOutputPaths

	return logger, err
}

//ViperToZap - configuration adapter.
/* handling viper's "feature"
viper lowercase all keys, while zap configuration is case sensitive
so this part of code is written
*/
func ViperToZap(bs []byte) []byte {
	r := strings.NewReplacer(
		"encoderconfig", "encoderConfig",
		"outputpaths", "outputPaths",
		"erroroutputpaths", "errorOutputPaths",
		"initialfields", "initialFields",
		"levelencoder", "levelEncoder",
		"levelkey", "levelKey",
		"messagekey", "messageKey",
	)
	return []byte(r.Replace(string(bs)))
}
