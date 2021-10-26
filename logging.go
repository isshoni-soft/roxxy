package roxxy

import (
	"fmt"
	"github.com/isshoni-soft/safe-channel"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var loggers []Logger
var logFileChannel *safe_channel.SafeStringChannel
var logFileCore string

var defaultLogger = NewLogger("logger", 16)
var dateLayout = "01-02-2006|15:04:05"
var logFileName = "-" + time.Now().Format(dateLayout) + ".log"
var logFileEnabled = false

type Logger struct {
	loggerChannel *safe_channel.SafeStringChannel

	prefix string
	storageIndex int
}

func InitLogfile(filePrefix string, fileName string) {
	if logFileEnabled {
		return
	}

	logFileCore = fileName
	logFileName = filepath.Join(filePrefix, fileName + logFileName)
	logFileChannel = safe_channel.NewSafeStringChannel(5)

	fixLogFileNameCollisions()

	logFileEnabled = true

	go logFileTick()
}

func GetLogger() *Logger {
	return defaultLogger
}

func NewLogger(prefix string, buffer int) *Logger {
	result := new(Logger)
	result.prefix =  "[" + time.Now().Format(dateLayout) +"]: " + prefix + "| "
	result.loggerChannel = safe_channel.NewSafeStringChannel(buffer)
	result.storageIndex = len(loggers)

	loggers = append(loggers, *result)

	go result.loggerTick()

	return result
}

func (l *Logger) loggerTick() {
	for str := range l.loggerChannel.Channel() {
		safe_channel.RunMain(func() {
			fmt.Println(str)
		})
	}
}

func (l *Logger) Shutdown() {
	l.loggerChannel.WaitForClose()

	length := len(loggers)

	if length > 1 {
		loggers = append(loggers[:l.storageIndex], loggers[l.storageIndex + 1])
	} else {
		loggers = []Logger{}
	}
}

func ShutdownLogging() {
	for _, logger := range loggers {
		if logger.Closed() {
			continue
		}

		logger.Shutdown()
	}

	if logFileEnabled {
		logFileChannel.WaitForClose()
	}
}

func (l *Logger) SetPrefix(str string) {
	l.prefix = str
}

func (l *Logger) Format(str ...string) (result string) {
	result = l.prefix

	for _, s := range str {
		result = result + s
	}

	return
}

func (l *Logger) Log(str ...string) {
	fmt.Println("queue log: " + strings.Join(str, " "))
	l.loggerChannel.Offer(l.Format(str...))
}

func (l *Logger) Closed() bool {
	return l.loggerChannel.Closed()
}

func fixLogFileNameCollisions() {
	num := 1

	for _, err := os.Stat(logFileName); os.IsExist(err); {
		logFileName = logFileCore + "-" + time.Now().Format(dateLayout) + "-" + strconv.FormatInt(int64(num), 10) + ".log"
		num++
	}
}

func logFileTick() {
	f, err := os.Create(logFileName)

	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(logFileName), os.ModePerm)

		if err != nil {
			panic(err)
		}

		f, err = os.Create(logFileName)
	} else {
		panic(err)
	}

	defer func(f *os.File) {
		err := f.Close()

		if err != nil {
			panic(err)
		}
	}(f)

	logFileChannel.ForEach(func(str string) {
		_, err := f.WriteString(str + "\n")
		if err != nil {
			return
		}
	})
}
