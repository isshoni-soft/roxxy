package roxxy

import (
	"fmt"
)

type Logger struct {
	messageQueue chan string
	shutdownThread chan bool
	shutdownAwk chan bool
	prefix string
	running bool
}

func NewLogger(prefix string) *Logger {
	result := &Logger{
		messageQueue: make(chan string, 10),
		shutdownThread: make(chan bool),
		shutdownAwk: make(chan bool),
		prefix: prefix,
		running: true,
	}

	isInit := make(chan bool, 1)
	go result.tick(isInit)
	<- isInit

	return result
}

func (l *Logger) Format(str ...string) string {
	result := l.prefix

	for _, s := range str {
		result = result + s
	}

	return result
}

func (l *Logger) Log(str ...string) {
	if !l.running {
		return
	}

	l.messageQueue <- l.Format(str...)
}

func (l *Logger) Shutdown() {
	if !l.running {
		return
	}

	l.shutdownThread <- true
	l.running = false

	<- l.shutdownAwk
}

func (l *Logger) Running() bool {
	return l.running
}

func (l *Logger) tick(isInit chan bool) {
	isInit <- true

	for {
		select {
		case message := <- l.messageQueue:
			fmt.Println(message)
		case <- l.shutdownThread:
			for {
				select {
				case message := <- l.messageQueue:
					fmt.Println(message)
				default:
					l.shutdownAwk <- true
					return
				}
			}
		}
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
