package test

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/natefinch/lumberjack"
	glog "github.com/nitsugaro/glogger"
	"github.com/sirupsen/logrus"
)

var FOLDER = "./logs/"
var FILENAME = "app.log"

func TestAddTable(t *testing.T) {
	r := gin.Default()

	keys := glog.Init(&glog.GoLoggerConf{
		Folder:              FOLDER,
		FileName:            FILENAME,
		TransactionIdHeader: "x-transaction-id",
		TransactionIdKey:    "transaction_id",
		LoggerKey:           "glogger",
		Lumberjack: &lumberjack.Logger{
			Filename:   FOLDER + FILENAME,
			MaxSize:    1,
			MaxAge:     10,
			MaxBackups: 10,
		},
		GinConfig: &glog.GinConfig{
			Engine:      r,
			EndpointApi: "/logs",
			MaxLogsApi:  500,
		},
		Level: logrus.DebugLevel,
	})

	go r.Run(":8080")

	t.Run("Log File Created", func(t *testing.T) {
		files := glog.GetLogFiles(keys)
		if files == nil || len(files) != 1 {
			t.Errorf("Expected Log File created")
		}
	})

	transactionId := uuid.NewString()
	beginTime := time.Now().Add(time.Minute * -1)
	endTime := time.Now().Add(time.Minute)

	t.Run("Logs Are Empty", func(t *testing.T) {
		logs := glog.ReadLogs(&glog.RequestLogs{BeginTime: &beginTime, EndTime: &endTime, TransactionId: transactionId}, keys)
		if logs == nil || len(logs) != 0 {
			t.Error("Expected Logs to be an empty list, but got", logs)
		}
	})

	logger := glog.Glob.WithField("transaction_id", transactionId)
	logger.Info("log 1")
	logger.Info("log 2")
	logger.Info("log 3")

	t.Run("Have 3 logs with transactionId", func(t *testing.T) {
		logs := glog.ReadLogs(&glog.RequestLogs{BeginTime: &beginTime, EndTime: &endTime, TransactionId: transactionId}, keys)
		if logs == nil || len(logs) != 3 {
			t.Error("Expected to get 3 logs: ", logs)
		}
	})

	logger.Info("log 3")

	t.Run("Query Filter Works 1", func(t *testing.T) {
		logs := glog.ReadLogs(&glog.RequestLogs{BeginTime: &beginTime, EndTime: &endTime, Query: &glog.Query{Qtype: "sw", Field: "msg", Value: "log 3"}}, keys)
		if logs == nil || len(logs) != 2 {
			t.Error("Expected expression 'msg sw log 3' = two logs and got: ", logs)
		}
	})

	logger.WithField("data", gin.H{"value": gin.H{"submsg": "This is the end"}}).Info("This is the begin!")

	t.Run("Query Filter Works 2", func(t *testing.T) {
		logs := glog.ReadLogs(&glog.RequestLogs{BeginTime: &beginTime, EndTime: &endTime, Query: &glog.Query{Qtype: "ew", Field: "data/value/submsg", Value: "end"}}, keys)
		if logs == nil || len(logs) != 1 {
			t.Error("Expected expression 'data/value/submsg ew end' = one log and got: ", logs)
		}
	})

	os.RemoveAll(FOLDER)
}
