package glog

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

var conf *GoLoggerConf
var Glob *logrus.Logger

func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			if ve, ok := err.Err.(validator.ValidationErrors); ok {
				for _, fe := range ve {
					c.JSON(400, gin.H{"error": fe.Field() + " doesn't has a valid format"})
					return
				}
			}

			c.JSON(400, gin.H{"error": "Bad Request"})
		}
	}
}

func Init(config *GoLoggerConf) *Keys {
	conf = config
	Glob = logrus.New()

	logrus.SetLevel(config.Level)

	Glob.SetFormatter(DefaultFormatter{&logrus.JSONFormatter{}})

	multiWriter := io.MultiWriter(os.Stdout, config.Lumberjack)
	Glob.SetOutput(multiWriter)

	var keys *Keys
	if conf.Validator == nil {
		keys = conf.defaultValidator()
	}

	if conf.GinConfig != nil {
		initGin()
	}

	return keys
}
