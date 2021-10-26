package roxxy

import (
	"fmt"
	"sync"
)

type Logger struct {
	mu sync.Mutex
	stringChannel *safechannel.SafeStringChannel
	prefix string
}

func NewLogger(prefix string) *Logger {
	result := &Logger {
		stringChannel: safechannel.NewSafeStringChannel(10),
		prefix: prefix,
	}

	isInit := make(chan bool, 1)
	go result.tick(isInit)
	<-isInit

	return result
}

func (l *Logger) Log(str string) {
	defer l.mu.Unlock()
	l.mu.Lock()

	l.stringChannel.Offer(str)
}

func (l *Logger) Shutdown() {
	defer l.mu.Unlock()
	l.mu.Lock()

	l.stringChannel.WaitForClose()
}

func (l *Logger) tick(isInit chan bool) {
	fmt.Println("init logger tick thread")
	isInit <- true

	for str := range l.stringChannel.Channel() {
		safechannel.RunMain(func() {
			fmt.Println(str)
		})
	}
}

//
//import (
//	"fmt"
//	"github.com/isshoni-soft/safe-channel"
//	"os"
//	"path/filepath"
//	"strconv"
//	"sync"
//	"time"
//)
//
//var loggers []*Logger
//var logFileChannel *safe_channel.SafeStringChannel
//var logFileCore string
//
//var dateLayout = "01-02-2006|15:04:05"
//var logFileName = "-" + time.Now().Format(dateLayout) + ".log"
//var logFileEnabled = false
//
//type Logger struct {
//	mu sync.Mutex
//	loggerChannel *safe_channel.SafeStringChannel
//
//	prefix string
//	storageIndex int
//}
//
//func InitLogFile(filePrefix string, fileName string) {
//	if logFileEnabled {
//		return
//	}
//
//	logFileCore = fileName
//	logFileName = filepath.Join(filePrefix, fileName + logFileName)
//	logFileChannel = safe_channel.NewSafeStringChannel(5)
//
//	fixLogFileNameCollisions()
//
//	logFileEnabled = true
//
//	go logFileTick()
//}
//
//func NewLogger(prefix string, buffer int) *Logger {
//	result := new(Logger)
//	result.prefix =  "[" + time.Now().Format(dateLayout) +"]: " + prefix + "| "
//	result.loggerChannel = safe_channel.NewSafeStringChannel(buffer)
//	result.storageIndex = len(loggers)
//
//	loggers = append(loggers, result)
//
//	ticked := make(chan bool, 1)
//	go result.loggerTick(ticked)
//	<-ticked
//
//	return result
//}
//
//func (l *Logger) loggerTick(hasTicked chan bool) {
//	safe_channel.RunMain(func() {
//		fmt.Println("logger ticking!")
//	}, true)
//
//	hasTicked <- true
//
//	for str := range l.loggerChannel.Channel() {
//		safe_channel.RunMain(func() {
//			fmt.Println(str)
//		}, true)
//	}
//}
//
//func (l *Logger) Shutdown() {
//	l.loggerChannel.WaitForClose()
//}
//
//func ShutdownLogging() {
//	for _, logger := range loggers {
//		if logger.Closed() {
//			continue
//		}
//
//		fmt.Println("Shutting down logger: " + logger.prefix)
//		logger.Shutdown()
//	}
//
//	if logFileEnabled {
//		logFileChannel.WaitForClose()
//	}
//}
//
//func (l *Logger) SetPrefix(str string) {
//	l.prefix = str
//}
//
//func (l *Logger) Format(str ...string) (result string) {
//	result = l.prefix
//
//	for _, s := range str {
//		result = result + s
//	}
//
//	return
//}
//
//func (l *Logger) Log(str ...string) {
//	l.mu.Lock()
//	fmt.Println("queuing log: " + l.Format(str...))
//	l.loggerChannel.Offer(l.Format(str...))
//	l.mu.Unlock()
//}
//
//func (l *Logger) Closed() bool {
//	return l.loggerChannel.Closed()
//}
//
//func fixLogFileNameCollisions() {
//	num := 1
//
//	for _, err := os.Stat(logFileName); os.IsExist(err); {
//		logFileName = logFileCore + "-" + time.Now().Format(dateLayout) + "-" + strconv.FormatInt(int64(num), 10) + ".log"
//		num++
//	}
//}
//
//func logFileTick() {
//	f, err := os.Create(logFileName)
//
//	if os.IsNotExist(err) {
//		err = os.MkdirAll(filepath.Dir(logFileName), os.ModePerm)
//
//		if err != nil {
//			panic(err)
//		}
//
//		f, err = os.Create(logFileName)
//	} else {
//		panic(err)
//	}
//
//	defer func(f *os.File) {
//		err := f.Close()
//
//		if err != nil {
//			panic(err)
//		}
//	}(f)
//
//	logFileChannel.ForEach(func(str string) {
//		_, err := f.WriteString(str + "\n")
//		if err != nil {
//			return
//		}
//	})
//}
