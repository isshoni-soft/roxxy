package roxxy

import (
	"fmt"
	"github.com/isshoni-soft/safe-channel"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var logFileChannel *safe_channel.SafeStringChannel

var defaultLogger = NewLogger("logger", 16)
var dateLayout = "01-02-2006|15:04:05"
var logFileName = "-" + time.Now().Format(dateLayout) + ".log"
var logFileEnabled = false

type Logger struct {
	loggerChannel *safe_channel.SafeStringChannel

	prefix string
}

func InitLogfile(filePrefix string, fileName string) {
	if logFileEnabled {
		return
	}

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

	go result.loggerTick()

	return result
}

func (l Logger) loggerTick() {
	l.loggerChannel.ForEach(func(str string) {
		fmt.Println(str)

		if logFileChannel != nil {
			logFileChannel.Offer(str) // now that we've logged the line lets queue it for adding to logfile
		}
	})
}

func (l Logger) Shutdown() {
	l.loggerChannel.WaitForClose()
}

func Shutdown() {
	logFileChannel.WaitForClose()
}

func (l Logger) SetPrefix(str string) {
	l.prefix = str
}

func (l Logger) Format(str ...string) (result string) {
	result = l.prefix

	for _, s := range str {
		result = result + s
	}

	return
}

func (l Logger) Log(str ...string) {
	l.loggerChannel.Offer(l.Format(str...))
}

func fixLogFileNameCollisions() {
	num := 1

	for _, err := os.Stat(logFileName); os.IsExist(err); {
		logFileName = "Sakura-" + time.Now().Format(dateLayout) + "-" + strconv.FormatInt(int64(num), 10) + ".log"
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
