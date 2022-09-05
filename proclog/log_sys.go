//go:build !windows
// +build !windows

package proclog

import (
	"errors"
	"fmt"
	"io"
	"log/syslog"
)

type SysLogger struct {
	NullLogger
	logWriter io.WriteCloser
}

func (that *SysLogger) Write(b []byte) (int, error) {
	if that.logWriter == nil {
		return 0, errors.New("not connect to syslog server")
	}
	return that.logWriter.Write(b)
}

func (that *SysLogger) Close() error {
	if that.logWriter == nil {
		return errors.New("not connect to syslog server")
	}
	return that.logWriter.Close()
}

func GetSyslogPriority(props map[string]string) syslog.Priority {
	logLevel := syslog.LOG_NOTICE
	if value, ok := props["syslog_priority"]; ok {
		logLevel = ToSyslogLevel(value)
	}
	facility := syslog.LOG_LOCAL0
	if value, ok := props["syslog_facility"]; ok {
		facility = ToSyslogFacility(value)
	}
	return logLevel | facility
}

// NewSysLogger 获取系统syslog的对象
func NewSysLogger(name string, props map[string]string) *SysLogger {
	priority := GetSyslogPriority(props)
	tag := name
	if value, ok := props["syslog_tag"]; ok {
		tag = value
	}
	writer, err := syslog.New(priority, tag)
	logger := &SysLogger{}
	if err == nil {
		logger.logWriter = writer
	}
	return logger
}

// NewRemoteSysLogger 获取远程系统日志的对象
func NewRemoteSysLogger(name string, config string, props map[string]string) *SysLogger {
	if len(config) <= 0 {
		return NewSysLogger(name, props)
	}

	protocol, host, port, err := ParseSysLogConfig(config)
	if err != nil {
		return NewSysLogger(name, props)
	}

	priority := GetSyslogPriority(props)
	tag := name
	if value, ok := props["syslog_tag"]; ok {
		tag = value
	}

	writer, err := syslog.Dial(protocol, fmt.Sprintf("%s:%d", host, port), priority, tag)
	logger := &SysLogger{}
	if writer != nil && err == nil {
		logger.logWriter = writer
	} else {
		logger.logWriter = NewBackendSysLogWriter(protocol, fmt.Sprintf("%s:%d", host, port), priority, name)
	}
	return logger
}
