package processes

import "github.com/moqsien/processes/proclog"

// 创建标准输出日志
func (that *ProcessPlus) CreateStdoutLogger() proclog.Logger {
	logFile := that.GetStdoutLogfile()
	maxBytes := int64(that.StdoutLogFileMaxBytes)
	backups := that.StdoutLogFileBackups

	props := make(map[string]string)
	return proclog.NewLogger(that.Name, logFile, proclog.NewNullLocker(), maxBytes, backups, props)
}

// 创建标准错误日志
func (that *ProcessPlus) CreateStderrLogger() proclog.Logger {
	logFile := that.GetStderrLogfile()
	maxBytes := int64(that.StderrLogFileMaxBytes)
	backups := that.StderrLogFileBackups

	props := make(map[string]string)
	return proclog.NewLogger(that.Name, logFile, proclog.NewNullLocker(), maxBytes, backups, props)
}
