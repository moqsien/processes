package proclog

import (
	"errors"
	"log/syslog"
	"strconv"
	"strings"
)

func SplitFileNames(fileNames string) []string {
	fileNameList := strings.Split(fileNames, ",")
	for i, f := range fileNameList {
		fileNameList[i] = strings.TrimSpace(f)
	}
	return fileNameList
}

func ToSyslogLevel(logLevel string) syslog.Priority {
	switch strings.ToUpper(logLevel) {
	case "EMERG", "LOG_EMERG":
		return syslog.LOG_EMERG
	case "ALERT", "LOG_ALERT":
		return syslog.LOG_ALERT
	case "CRIT", "CRITICAL", "LOG_CRIT", "LOG_CRITICAL":
		return syslog.LOG_CRIT
	case "ERR", "ERROR", "LOG_ERR", "LOG_ERROR":
		return syslog.LOG_ERR
	case "WARNING", "WARN", "LOG_WARNING", "LOG_WARN":
		return syslog.LOG_WARNING
	case "NOTICE", "LOG_NOTICE":
		return syslog.LOG_NOTICE
	case "INFO", "LOG_INFO":
		return syslog.LOG_INFO
	case "DEBUG", "LOG_DEBUG":
		return syslog.LOG_DEBUG
	default:
		return syslog.LOG_INFO
	}
}

func ToSyslogFacility(facility string) syslog.Priority {
	switch strings.ToUpper(facility) {
	case "KERN", "KERNEL", "LOG_KERN", "LOG_KERNEL":
		return syslog.LOG_KERN
	case "USER", "LOG_USER":
		return syslog.LOG_USER
	case "MAIL", "LOG_MAIL":
		return syslog.LOG_MAIL
	case "DAEMON", "LOG_DAEMON":
		return syslog.LOG_DAEMON
	case "AUTH", "LOG_AUTH":
		return syslog.LOG_AUTH
	case "SYSLOG", "LOG_SYSLOG":
		return syslog.LOG_SYSLOG
	case "LPR", "LOG_LPR":
		return syslog.LOG_LPR
	case "NEWS", "LOG_NEWS":
		return syslog.LOG_NEWS
	case "UUCP", "LOG_UUCP":
		return syslog.LOG_UUCP
	case "CRON", "LOG_CRON":
		return syslog.LOG_CRON
	case "AUTHPRIV", "LOG_AUTHPRIV":
		return syslog.LOG_AUTHPRIV
	case "FTP", "LOG_FTP":
		return syslog.LOG_FTP
	case "LOCAL0", "LOG_LOCAL0":
		return syslog.LOG_LOCAL0
	case "LOCAL1", "LOG_LOCAL1":
		return syslog.LOG_LOCAL1
	case "LOCAL2", "LOG_LOCAL2":
		return syslog.LOG_LOCAL2
	case "LOCAL3", "LOG_LOCAL3":
		return syslog.LOG_LOCAL3
	case "LOCAL4", "LOG_LOCAL4":
		return syslog.LOG_LOCAL4
	case "LOCAL5", "LOG_LOCAL5":
		return syslog.LOG_LOCAL5
	case "LOCAL6", "LOG_LOCAL6":
		return syslog.LOG_LOCAL6
	case "LOCAL7", "LOG_LOCAL7":
		return syslog.LOG_LOCAL7
	default:
		return syslog.LOG_LOCAL0

	}
}

// parse the configuration for syslog, it should be in following format:
// [protocol:]host[:port]
//
// - protocol, could be tcp or udp, assuming udp as default
// - port, if missing, by default for tcp is 6514 and for udp - 514
func ParseSysLogConfig(config string) (protocol string, host string, port int, err error) {
	fields := strings.Split(config, ":")
	host = ""
	protocol = ""
	port = 0
	err = nil
	switch len(fields) {
	case 1:
		host = fields[0]
		port = 514
		protocol = "udp"
	case 2:
		switch fields[0] {
		case "tcp":
			host = fields[1]
			protocol = "tcp"
			port = 6514
		case "udp":
			host = fields[1]
			protocol = "udp"
			port = 514
		default:
			protocol = "udp"
			host = fields[0]
			port, err = strconv.Atoi(fields[1])
			if err != nil {
				return
			}
		}
	case 3:
		protocol = fields[0]
		host = fields[1]
		port, err = strconv.Atoi(fields[2])
		if err != nil {
			return
		}
	default:
		err = errors.New("invalid format")
	}
	return
}
