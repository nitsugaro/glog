package glog

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type GinConfig struct {
	Engine      *gin.Engine
	EndpointApi string `json:"endpointApi"`
	MaxLogsApi  int    `json:"max_logs_api"`
}

// Use with TransactionIdMiddlewareGin middleware with dynamic config for each request. If there isn't different loggers for request, instead use GLogger.
func GetLoggerGinCtx(c *gin.Context) *logrus.Entry {
	logger, ok := c.Get(conf.LoggerKey)
	if ok {
		return logger.(*logrus.Entry)
	}

	return nil
}

// Set a transactionId to logger in gin ctx. All entry logs will have that field for trace.
func TransactionIdMiddlewareGin(c *gin.Context) {
	transactionID := c.Request.Header.Get(conf.TransactionIdHeader)

	if transactionID == "" {
		transactionID = uuid.New().String()
	}

	logEntry := Glob.WithField(conf.TransactionIdKey, transactionID)

	logEntry.Logger.SetLevel(conf.Level)

	c.Set(conf.LoggerKey, logEntry)

	c.Writer.Header().Set(conf.TransactionIdHeader, transactionID)

	c.Next()
}

func getLogsGin(c *gin.Context) {
	log := GetLoggerGinCtx(c)
	if log == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Error"})
		return
	}

	keys := &Keys{
		apiKey:    c.GetHeader("x-api-key"),
		apiSecret: c.GetHeader("x-api-secret"),
	}
	if !conf.Validator.Validate(keys) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unathorized"})
		return
	}

	params := &RequestLogs{}
	if err := c.ShouldBindJSON(params); err != nil {
		log.Error(err)
		c.Error(err)
		return
	}

	logs := ReadLogs(params, keys)

	length := len(logs)

	c.JSON(http.StatusOK, &BasePaginatedResult{
		ResultCount: length,
		Stop:        length != conf.MaxLogsApi,
		Result:      logs,
	})
}

func deleteAllLogsGin(c *gin.Context) {
	keys := &Keys{
		apiKey:    c.GetHeader("x-api-key"),
		apiSecret: c.GetHeader("x-api-secret"),
	}
	if !conf.Validator.Validate(keys) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unathorized"})
		return
	}

	logger := GetLoggerGinCtx(c)
	if logger != nil {
		logger.Info("Logs restarted")
	}

	ResetLogs(keys)
}

func initGin() {
	conf.Engine.Use(errorHandler())
	conf.Engine.POST(conf.EndpointApi, TransactionIdMiddlewareGin, getLogsGin)
	conf.Engine.DELETE(conf.EndpointApi, TransactionIdMiddlewareGin, deleteAllLogsGin)
}
