package processes

import (
	"fmt"
	"syscall"
	"time"

	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gtime"
)

// Info 进程的运行状态
type Info struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Start         int    `json:"start"`
	Stop          int    `json:"stop"`
	Now           int    `json:"now"`
	State         int    `json:"state"`
	StateName     string `json:"statename"`
	SpawnErr      string `json:"spawnerr"`
	ExitStatus    int    `json:"exitstatus"`
	Logfile       string `json:"logfile"`
	StdoutLogfile string `json:"stdout_logfile"`
	StderrLogfile string `json:"stderr_logfile"`
	Pid           int    `json:"pid"`
}

// GetProcessInfo 获取进程的详情
func (that *ProcessPlus) GetProcessInfo() *Info {
	return &Info{
		Name:          that.Name,
		Description:   that.GetDescription(),
		Start:         int(that.StartTime.Unix()),
		Stop:          int(that.StopTime.Unix()),
		Now:           int(time.Now().Unix()),
		State:         int(that.State),
		StateName:     that.State.ToString(),
		SpawnErr:      "",
		ExitStatus:    that.GetExitStatus(),
		Logfile:       that.GetStdoutLogfile(),
		StdoutLogfile: that.GetStdoutLogfile(),
		StderrLogfile: that.GetStderrLogfile(),
		Pid:           that.Pid()}

}

// GetDescription 获取进程描述
func (that *ProcessPlus) GetDescription() string {
	that.Lock.RLock()
	defer that.Lock.RUnlock()
	if that.State == Running {
		seconds := int(time.Now().Sub(that.StartTime).Seconds())
		minutes := seconds / 60
		hours := minutes / 60
		days := hours / 24
		if days > 0 {
			return fmt.Sprintf("pid %d, uptime %d days, %d:%02d:%02d", that.Cmd.Process.Pid, days, hours%24, minutes%60, seconds%60)
		}
		return fmt.Sprintf("pid %d, uptime %d:%02d:%02d", that.Cmd.Process.Pid, hours%24, minutes%60, seconds%60)
	} else if that.State != Stopped {
		return gtime.New(that.StopTime).String()
	}
	return ""
}

// GetExitStatus 获取进程退出状态
func (that *ProcessPlus) GetExitStatus() int {
	that.Lock.RLock()
	defer that.Lock.RUnlock()

	if that.State == Exited || that.State == Suspend {
		if that.ProcessState == nil {
			return 0
		}
		status, ok := that.ProcessState.Sys().(syscall.WaitStatus)
		if ok {
			return status.ExitStatus()
		}
	}
	return 0
}

// GetStdoutLogfile 获取标准输出将要写入的日志文件
func (that *ProcessPlus) GetStdoutLogfile() string {
	fileName := "/dev/null"
	if len(that.StdoutLogfile) > 0 {
		fileName = that.StdoutLogfile
	}
	expandFile := gfile.RealPath(fileName)
	return expandFile
}

// GetStderrLogfile 获取标准错误将要写入的日志文件
func (that *ProcessPlus) GetStderrLogfile() string {
	fileName := "/dev/null"
	if len(that.StderrLogfile) > 0 {
		fileName = that.StdoutLogfile
	}
	expandFile := gfile.RealPath(fileName)
	return expandFile
}

// GetStatus 获取进程当前状态
func (that *ProcessPlus) GetStatus() string {
	if that.ProcessState.Exited() {
		return that.ProcessState.String()
	}
	return "running"
}

// Pid 获取进程pid，返回0表示进程未启动
func (that *ProcessPlus) Pid() int {
	if (Failure & that.State) != 0 {
		return 0
	}
	return that.Process.Pid
}
