package glog

import (
	"github.com/google/uuid"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

type Keys struct {
	apiKey    string
	apiSecret string
}

type GoLoggerConf struct {
	Folder              string             `json:"folder"`
	FileName            string             `json:"file_name"`
	TransactionIdHeader string             `json:"transaction_id_header"`
	TransactionIdKey    string             `json:"transaction_id_key"`
	LoggerKey           string             `json:"logger_key"`
	Lumberjack          *lumberjack.Logger `json:"lumberjack"`
	Level               logrus.Level       `json:"level"`
	*GinConfig

	Validator Validator
}

func (gConf *GoLoggerConf) defaultValidator() *Keys {
	apiKey := uuid.NewString()
	apiSecret := uuid.NewString() + uuid.NewString()
	gConf.Validator = &SimpleValidator{apiKey: apiKey, apiSecret: apiSecret}

	Glob.Infof("GLogger credentials apiKey: %s, apiSecret: %s", apiKey, apiSecret)

	return &Keys{apiKey: apiKey, apiSecret: apiSecret}
}
