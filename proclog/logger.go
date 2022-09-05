package proclog

import (
	"io"
	"strings"
	"sync"
)

// Logger 日志接口
type Logger interface {
	io.WriteCloser
	SetPid(pid int)
	ReadLog(offset int64, length int64) (string, error)
	ReadTailLog(offset int64, length int64) (string, int64, bool, error)
	ClearCurLogFile() error
	ClearAllLogFile() error
}

// 创建日志对象
func CreateLogger(programName string,
	logFileName string,
	locker sync.Locker,
	maxBytes int64,
	backups int,
	props map[string]string) Logger {

	if logFileName == "/dev/stdout" {
		return NewStdoutLogger()
	}
	if logFileName == "/dev/stderr" {
		return NewStderrLogger()
	}
	if logFileName == "/dev/null" {
		return NewNullLogger()
	}

	if logFileName == "syslog" {
		return NewSysLogger(programName, props)
	}
	if strings.HasPrefix(logFileName, "syslog") {
		fields := strings.Split(logFileName, "@")
		fields[0] = strings.TrimSpace(fields[0])
		fields[1] = strings.TrimSpace(fields[1])
		if len(fields) == 2 && fields[0] == "syslog" {
			return NewRemoteSysLogger(programName, fields[1], props)
		}
	}

	if len(logFileName) > 0 {
		return NewFileLogger(logFileName, maxBytes, backups, locker)
	}
	return NewNullLogger()
}

// NewLogger 新建日志对象
func NewLogger(programName string,
	logFileNames string,
	locker sync.Locker,
	maxBytes int64,
	backups int,
	props map[string]string) Logger {

	files := SplitFileNames(logFileNames)
	loggers := make([]Logger, 0)
	for i, f := range files {
		var lg Logger
		if i == 0 {
			lg = CreateLogger(programName,
				f,
				locker,
				maxBytes,
				backups,
				props)
		} else {
			lg = CreateLogger(programName,
				f,
				NewNullLocker(),
				maxBytes,
				backups,
				props)
		}
		loggers = append(loggers, lg)
	}
	return NewCompositeLogger(loggers)
}

/*
复合日志类型
*/
type CompositeLogger struct {
	lock    sync.Mutex
	loggers []Logger
}

func (that *CompositeLogger) Write(p []byte) (n int, err error) {
	that.lock.Lock()
	defer that.lock.Unlock()

	for i, logger := range that.loggers {
		if i == 0 {
			n, err = logger.Write(p)
		} else {
			_, _ = logger.Write(p)
		}
	}
	return
}

func (that *CompositeLogger) Close() (err error) {
	that.lock.Lock()
	defer that.lock.Unlock()

	for i, logger := range that.loggers {
		if i == 0 {
			err = logger.Close()
		} else {
			_ = logger.Close()
		}
	}
	return
}

func (that *CompositeLogger) SetPid(pid int) {
	that.lock.Lock()
	defer that.lock.Unlock()

	for _, logger := range that.loggers {
		logger.SetPid(pid)
	}
}

// ReadLog read log data from first logger in CompositeLogger pool
func (that *CompositeLogger) ReadLog(offset int64, length int64) (string, error) {
	return that.loggers[0].ReadLog(offset, length)
}

// ReadTailLog tail the log data from first logger in CompositeLogger pool
func (that *CompositeLogger) ReadTailLog(offset int64, length int64) (string, int64, bool, error) {
	return that.loggers[0].ReadTailLog(offset, length)
}

// ClearCurLogFile clear the first logger file in CompositeLogger pool
func (that *CompositeLogger) ClearCurLogFile() error {
	return that.loggers[0].ClearCurLogFile()
}

// ClearAllLogFile clear all the files of first logger in CompositeLogger pool
func (that *CompositeLogger) ClearAllLogFile() error {
	return that.loggers[0].ClearAllLogFile()
}

func (that *CompositeLogger) AddLogger(logger Logger) {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.loggers = append(that.loggers, logger)
}

func (that *CompositeLogger) RemoveLogger(logger Logger) {
	that.lock.Lock()
	defer that.lock.Unlock()
	for i, t := range that.loggers {
		if t == logger {
			that.loggers = append(that.loggers[:i], that.loggers[i+1:]...)
			break
		}
	}
}

func NewCompositeLogger(loggers []Logger) *CompositeLogger {
	return &CompositeLogger{loggers: loggers}
}
