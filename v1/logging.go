package roxxy_v1

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const defaultTimeFormat = "[01/02/2006|03:04:05.00]: "

var writerRunning = false
var openFiles = make(map[string]os.File)
var loggersByOpenFile = make(map[string][]Logger)
var writerChannel = make(chan LoggerEntry)

type LoggerEntry struct {
	message string
	file    *os.File
}

type Logger struct {
	messageQueue   chan string
	shutdownThread chan bool
	shutdownAwk    chan bool
	prefix         string
	fileSuffix     string
	fileKey        string
	timeFormat     string
	running        bool
	timestamp      bool
	logFile        *os.File
}

func NewFileLogger(prefix string, filePrefix string, fileName string, fileSuffix string) *Logger {
	logger := NewLogger(prefix)
	logger.fileSuffix = fileSuffix + ".log"

	logger.StartLoggingToFile(filePrefix, fileName)

	return logger
}

func NewLoggerWithoutTimestamp(prefix string) *Logger {
	logger := NewLogger(prefix)
	logger.timestamp = false

	return logger
}

func NewLogger(prefix string) *Logger {
	result := &Logger{
		messageQueue:   make(chan string, 10),
		shutdownThread: make(chan bool),
		shutdownAwk:    make(chan bool),
		prefix:         prefix,
		timeFormat:     defaultTimeFormat,
		running:        true,
		timestamp:      true,
	}

	isInit := make(chan bool, 1)
	go result.tick(isInit)
	<-isInit

	return result
}

func tickLoggerFiles(started chan bool) {
	started <- true

	for entry := range writerChannel {
		_, err := entry.file.WriteString(entry.message)

		if err != nil {
			panic(err)
		}
	}
}

func fixLogFileNameCollisions(logFileName string, logFileCore string, logFileSuffix string) string {
	num := 1

	for ok, _ := exists(logFileName); ok; ok, _ = exists(logFileName) {
		logFileName = logFileCore + "-" + strconv.FormatInt(int64(num), 10) + logFileSuffix
		num++
	}

	return logFileName
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func removeIndexFromLoggersByOpenFile(key string, index int) {
	if len(loggersByOpenFile[key]) == 1 {
		loggersByOpenFile[key][0].logFile.Close()
		delete(loggersByOpenFile, key)
	} else {
		loggersByOpenFile[key] = append(loggersByOpenFile[key][:index], loggersByOpenFile[key][index+1])
	}
}

func (l *Logger) StartLoggingToFile(filePrefix string, fileName string) {
	if l.logFile != nil {
		return
	}

	logFileName := filepath.Join(filePrefix, fileName+l.fileSuffix)
	l.fileKey = logFileName

	if val, ok := openFiles[l.fileKey]; ok && writerRunning {
		l.logFile = &val
		return
	}

	logFileName = fixLogFileNameCollisions(logFileName, fileName, l.fileSuffix)

	if _, err := os.Stat(logFileName); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(logFileName), os.ModePerm)

		if err != nil {
			panic(err)
		}

		f, e := os.Create(logFileName)

		if e != nil {
			panic(e)
		}

		l.logFile = f

		openFiles[l.fileKey] = *f

		if val, ok := loggersByOpenFile[l.fileKey]; ok {
			val = append(val, *l)
		} else {
			loggersByOpenFile[l.fileKey] = []Logger{*l}
		}

		if !writerRunning {
			writerRunning = true

			started := make(chan bool)
			go tickLoggerFiles(started)
			<-started
		}
	} else {
		panic(err)
	}
}

func (l *Logger) Format(str ...string) string {
	result := l.prefix + " "

	for _, s := range str {
		result = result + s
	}

	return result
}

func (l *Logger) Log(str ...string) {
	if !l.running {
		return
	}

	message := l.Format(str...)

	l.messageQueue <- message

	if l.logFile != nil {
		writerChannel <- LoggerEntry{
			message: message + "\n",
			file:    l.logFile,
		}
	}
}

func (l *Logger) Shutdown() {
	if !l.running {
		return
	}

	l.shutdownThread <- true
	l.running = false

	if l.logFile != nil {
		delete(openFiles, l.fileKey)

		var remove int
		for index, logger := range loggersByOpenFile[l.fileKey] {
			if *l == logger {
				remove = index
			}
		}

		removeIndexFromLoggersByOpenFile(l.fileKey, remove)
	}

	<-l.shutdownAwk
}

func (l *Logger) Running() bool {
	return l.running
}

func (l *Logger) handleMessage(message string) {
	if l.timestamp {
		message = time.Now().Format(l.timeFormat) + message
	}

	fmt.Println(message)
}

func (l *Logger) tick(isInit chan bool) {
	isInit <- true

	for {
		select {
		case message := <-l.messageQueue:
			l.handleMessage(message)
		case <-l.shutdownThread:
			for {
				select {
				case message := <-l.messageQueue:
					l.handleMessage(message)
				default:
					l.shutdownAwk <- true
					return
				}
			}
		}
	}
}
