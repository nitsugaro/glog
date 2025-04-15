package glog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

const UTC_FORMAT = "2006-01-02T15:04:05Z"

type LogFile struct {
	File      fs.DirEntry
	Timestamp string
}

type LogEntry struct {
	Level         string                 `json:"level"`
	Time          time.Time              `json:"time"`
	Msg           string                 `json:"msg"`
	TransactionId string                 `json:"transaction_id,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

type Query struct {
	Qtype string `json:"qtype" binding:"oneof=eq sw ew co" msg:"cualqueiraaa"`
	Value any    `json:"value" binding:"required"`
	Field string `json:"field" binding:"required"`
}

type RequestLogs struct {
	BeginTime     *time.Time `json:"begin_time" time_format:"2006-01-02T15:04:05Z" binding:"omitempty"`
	EndTime       *time.Time `json:"end_time" time_format:"2006-01-02T15:04:05Z" binding:"omitempty"`
	Query         *Query     `json:"query" binding:"omitempty"`
	TransactionId string     `json:"transaction_id" binding:"omitempty"`
}

var re = regexp.MustCompile(`app-(\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}\.\d{3})\.log`)

func passQueryEval(query *Query, logEntry *LogEntry) bool {
	if query == nil {
		return true
	}

	var currentVal any = *logEntry
	fields := strings.Split(query.Field, "/")

	for _, field := range fields {
		switch v := currentVal.(type) {
		case map[string]interface{}:
			next, ok := v[field]
			if !ok {
				return false
			}
			currentVal = next

		default:
			val := reflect.ValueOf(currentVal)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				found := false
				valType := val.Type()
				for i := 0; i < valType.NumField(); i++ {
					fieldInfo := valType.Field(i)
					tag := fieldInfo.Tag.Get("json")
					tag = strings.Split(tag, ",")[0] // por si tiene `omitempty` u otros
					if tag == field {
						fieldVal := val.Field(i)
						if !fieldVal.IsValid() {
							return false
						}
						currentVal = fieldVal.Interface()
						found = true
						break
					}
				}
				if !found {
					return false
				}
			} else {
				return false
			}
		}
	}

	// Comparación según qtype
	strVal := currentVal
	switch query.Qtype {
	case "eq":
		return strVal == query.Value
	case "sw":
		return strings.HasPrefix(strVal.(string), query.Value.(string))
	case "ew":
		return strings.HasSuffix(strVal.(string), query.Value.(string))
	case "co":
		return strings.Contains(strVal.(string), query.Value.(string))
	default:
		return false
	}
}

func parseLog(line string) (*LogEntry, error) {
	var entry LogEntry
	err := json.Unmarshal([]byte(line), &entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func findLogsInRange(filePath string, queryParams *RequestLogs, maxLogs *int) ([]*LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		Glob.Errorf("Error open file '%s': %s", filePath, err)
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()

	var logs []*LogEntry
	left, right := int64(0), fileSize
	foundPosition := int64(-1)

	for left <= right {
		mid := (left + right) / 2

		_, err := file.Seek(mid, 0)
		if err != nil {
			return nil, err
		}

		reader := bufio.NewReader(file)
		_, _ = reader.ReadString('\n')

		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		entry, err := parseLog(strings.TrimSpace(line))
		if err != nil {
			continue
		}

		if entry.Time.Before(*queryParams.BeginTime) {
			left = mid + 1
		} else {
			foundPosition = mid
			right = mid - 1
		}
	}

	if foundPosition == -1 {
		return nil, nil
	}

	_, _ = file.Seek(foundPosition, 0)
	reader := bufio.NewScanner(file)

	for reader.Scan() {
		entry, err := parseLog(reader.Text())
		if err != nil {
			continue
		}

		if entry.Time.After(*queryParams.EndTime) {
			break
		}

		if (queryParams.TransactionId != "" && queryParams.TransactionId != entry.TransactionId) || !passQueryEval(queryParams.Query, entry) {
			continue
		}

		logs = append(logs, entry)
		(*maxLogs)++

		if *maxLogs >= conf.MaxLogsApi {
			break
		}
	}

	return logs, nil
}

func ReadLogs(params *RequestLogs, keys *Keys) []*LogEntry {
	files := GetLogFiles(keys)

	if files == nil {
		return nil
	}

	logs := []*LogEntry{}
	logsCount := 0

	for _, file := range files {
		//file rotation timestamp must be lower than beginTime
		if file.Timestamp != "" && params.BeginTime.Format(UTC_FORMAT) > file.Timestamp {
			fmt.Println(params.BeginTime.Format(UTC_FORMAT), file.Timestamp)
			continue
		}

		newLogs, err := findLogsInRange(conf.Folder+file.File.Name(), params, &logsCount)
		if err != nil {
			Glob.Errorf("Error trying to read logs in file '%s': error %s", file.File.Name(), err.Error())
			continue
		}

		logs = append(logs, newLogs...)

		if logsCount >= conf.MaxLogsApi {
			break
		}
	}

	return logs
}

func ResetLogs(keys *Keys) bool {
	if !conf.Validator.Validate(keys) {
		return false
	}

	conf.Lumberjack.Rotate()

	dir, err := os.ReadDir(conf.Folder)
	if err != nil {
		Glob.Errorf("Resetlogs error trying: %s", err.Error())
	}

	for _, entry := range dir {
		if entry.Name() != conf.FileName && !entry.IsDir() {
			err := os.Remove(conf.Folder + entry.Name())
			if err != nil {
				return false
			}
		}
	}

	return err != nil
}

// Returns a list of all logs file in folder. The name and timestamp from rotation.
func GetLogFiles(keys *Keys) []LogFile {
	if !conf.Validator.Validate(keys) {
		return nil
	}

	files, err := os.ReadDir(conf.Folder)

	if err != nil {
		Glob.Error("Error getting file name logs", err)
	}

	logFiles := []LogFile{}

	for _, file := range files {
		match := re.FindStringSubmatch(file.Name())
		timestamp := ""
		if len(match) > 1 {
			timestamp = match[1] + "Z"
		}

		logFiles = append(logFiles, LogFile{
			File:      file,
			Timestamp: timestamp,
		})
	}

	sort.Slice(logFiles, func(i int, j int) bool {
		return logFiles[i].Timestamp < logFiles[j].Timestamp
	})

	return logFiles
}
