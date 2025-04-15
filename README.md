# Library to generate default API for logs in JSON format

### Basic Configuration

```go
glog.Init(&glog.GoLoggerConf{
		Folder:              "/myfolder/path/",
		FileName:            "app.log",
		TransactionIdHeader: "x-transaction-id",
		TransactionIdKey:    "transaction_id",
		LoggerKey:           "glogger",
		Lumberjack: &lumberjack.Logger{
			Filename:   "/myfolder/path/app.log",
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
```

### Global Log

```go
glog.Glob.Info("My global log")
```

output
```json
{
	"level": "info",
	"msg": "My global log",
	"time": "2025-04-03T01:55:31Z"
}
```

### Gin Log Transaction

```go
func getLogsGin(c *gin.Context) {
	logger := GetLoggerGinCtx(c)
	if logger == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Error"})
		return
	}

	logger.Info("My log with Transaction Id")
	logger.WithFields("data", gin.H{ "value": 1 }).Info("My log with data")
}
```

output
```json
{
	"level": "info",
	"msg": "My log with Transaction Id",
	"time": "2025-04-03T01:55:31Z",
	"transaction_id": "4a07b398-7111-48a5-bbab-25bdfc8416ea"
}
{
	"data": {
		"value": 1
	},
	"level": "info",
	"msg": "My log with data",
	"time": "2025-04-03T01:55:31Z",
	"transaction_id": "4a07b398-7111-48a5-bbab-25bdfc8416ea"
}
```

#### Get Logs

Query Types: `eq`, `sw`, `ew`, `co`

```bash
curl --location 'http://<MY-HOST>/logs' \
--header 'x-api-key: <API-KEY>' \
--header 'x-api-secret: <API-SECRET>' \
--header 'Content-Type: application/json' \
--data '{
    "transaction_id?": "ea730c73-56b6-4436-8506-3393af1646e3",
    "begin_time": "2025-04-03T00:00:36Z",
    "end_time": "2025-04-03T10:00:36Z",
    "query?": {
        "qtype": "sw",
        "field": "msg",
        "value": "Reading"
    }
}'
```
output
```json
{
    "resultCount": 500,
    "stop": false,
    "result": [
        {
            "level": "info",
            "time": "2025-04-03T01:48:57Z",
            "msg": "Reading logs..."
        },
        {
            "level": "info",
            "time": "2025-04-03T01:53:15Z",
            "msg": "GLogger credentials apiKey: 6551a8da-0cd3-4061-9789-a038cbee40bd, apiSecret: 10e7b3e5-e9ff-4a0e-91b6-d3b91ae1ff972c6d6aa7-ac50-459f-ae6c-f5a77990bc0c"
        },
        {
            "level": "info",
            "time": "2025-04-03T01:53:19Z",
            "msg": "Reading logs...",
            "data": {
                "value": 1
            }
        },
		...
    ]
}
```

#### Delete Logs

```bash
curl --location --request DELETE 'http://<MY-HOST>/logs' \
--header 'x-api-key: <API-KEY>' \
--header 'x-api-secret: <API-SECRET>'
```

output
`status code 200`