package processes

import (
	"os"

	"github.com/gogf/gf/container/gmap"
	"github.com/moqsien/processes/utils"
)

type ProcSettings struct {
	Environment *gmap.StrStrMap // 环境变量

	AutoStart             bool        // 启动的时候自动该进程启动
	StartSecs             int         // 启动10秒后没有异常退出，就表示进程正常启动了，默认为1秒
	AutoReStart           AutoReStart // 程序退出后自动重启的规则,可选值：[unexpected,true,false]，默认为unexpected，表示进程意外杀死后才重启
	ExitCodes             []int       // 进程退出的code值
	StartRetries          int         // 启动失败自动重试次数，默认是3
	RestartPause          int         // 进程重启间隔秒数，默认是0，表示不间隔
	User                  string      // 用哪个用户启动进程，默认是父进程的所属用户
	Priority              int         // 进程启动优先级，默认999，值小的优先启动
	StdoutLogfile         string      // 日志文件，需要注意当指定目录不存在时无法正常启动，所以需要手动创建目录（supervisord 会自动创建日志文件）
	StdoutLogFileMaxBytes int         // stdout 日志文件大小，默认50MB
	StdoutLogFileBackups  int         // stdout 日志文件备份数，默认是10
	RedirectStderr        bool        // 把stderr重定向到stdout，默认false
	StderrLogfile         string      // 日志文件，进程启动后的标准错误写入该文件
	StderrLogFileMaxBytes int         // stderr 日志文件大小，默认50MB
	StderrLogFileBackups  int         // stderr 日志文件备份数，默认是10

	StopAsGroup              bool            // 默认为false,进程被杀死时，是否向这个进程组发送stop信号，包括子进程
	KillAsGroup              bool            // 默认为false，向进程组发送kill信号，包括子进程
	StopSignal               []string        // 结束进程发送的信号
	StopWaitSecs             int             // 发送结束进程的信号后等待的秒数
	KillWaitSecs             int             // 强制杀死进程等待秒数
	RestartWhenBinaryChanged bool            // 当进程的二进制文件有修改，是否需要重启,默认false
	Extend                   *gmap.AnyAnyMap // 扩展参数
}

/*
进程相关配置
*/

type Option func(p *ProcessPlus)

// SetName 设置进程名称
func ProcName(name string) Option {
	return func(p *ProcessPlus) {
		p.Name = name
	}
}

// SetProcPath 设置启动命令的path
func ProcPath(path string) Option {
	return func(p *ProcessPlus) {
		p.Path = path
		if len(p.Args) > 0 {
			p.Args[0] = path
		} else {
			p.Args = []string{path}
		}
	}
}

// SetProcArgs 设置启动命令的参数
func ProcArgs(args []string) Option {
	return func(p *ProcessPlus) {
		p.Args = append(p.Args, args...)
	}
}

// func (that *ProcessPlus) SetProcArgs(args []string) {
// 	that.Args = append(that.Args, args...)
// }

// SetProcEnvVar 设置进程专有的环境变量
func ProcEnvVar(key, value string) Option {
	return func(p *ProcessPlus) {
		p.Environment.Set(key, value)
	}
}

// func (that *ProcessPlus) SetProcEnvVar(key, value string) {
// 	that.Environment.Set(key, value)
// }

// SetProcEnvVarByMap 设置进程的环境变量，通过map的方式
func ProcEnvVarByMap(envs map[string]string) Option {
	return func(p *ProcessPlus) {
		p.Environment.Sets(envs)
	}
}

// func (that *ProcessPlus) SetProcEnvVarByMap(envs map[string]string) {
// 	that.Environment.Sets(envs)
// }

// SetProcStdoutLog 设置进程标准日志输出的文件
func ProcStdoutLog(file, maxBytes string, backups ...int) Option {
	return func(p *ProcessPlus) {
		p.StdoutLogfile = file
		p.StdoutLogFileMaxBytes = utils.GetBytes(maxBytes, 50*1024*1024)
		p.StdoutLogFileBackups = 10
		if len(backups) > 0 {
			p.StdoutLogFileBackups = backups[0]
		}
	}
}

// func (that *ProcessPlus) SetProcStdoutLog(file string, maxBytes string, backups ...int) {
// 	that.StdoutLogfile = file
// 	that.StdoutLogFileMaxBytes = utils.GetBytes(maxBytes, 50*1024*1024)
// 	that.StdoutLogFileBackups = 10
// 	if len(backups) > 0 {
// 		that.StdoutLogFileBackups = backups[0]
// 	}
// }

// SetProcRedirectStderr 设置错误输出与标准输出一起
func ProcRedirectStderr(r bool) Option {
	return func(p *ProcessPlus) {
		p.RedirectStderr = r
	}
}

// func (that *ProcessPlus) SetProcRedirectStderr(r bool) {
// 	that.RedirectStderr = r
// }

// ProcAutoReStart 设置进程自动重启的模式
func ProcAutoReStart(a AutoReStart) Option {
	return func(p *ProcessPlus) {
		p.AutoReStart = a
	}
}

// func (that *ProcessPlus) SetProcAutoReStart(a AutoReStart) {
// 	that.AutoReStart = a
// }

// SetProcExtraFiles 设置子进程从父进程继承的文件句柄
func ProcExtraFiles(fList []*os.File) Option {
	return func(p *ProcessPlus) {
		p.ExtraFiles = fList
	}
}

// func (that *ProcessPlus) SetProcExtraFiles(fileList []*os.File) {
// 	that.ExtraFiles = fileList
// }

// SetProcStopSignal 设置进程退出时发送的信号
func ProcStopSignal(sigs ...string) Option {
	return func(p *ProcessPlus) {
		p.StopSignal = sigs
	}
}

// func (that *ProcessPlus) SetProcStopSignal(sigs ...string) {
// 	that.StopSignal = sigs
// }

// SetProcDirectory 设置进程运行目录
func ProcDirectory(dir string) Option {
	return func(p *ProcessPlus) {
		p.Dir = dir
	}
}

// func (that *ProcessPlus) SetProcDirectory(dir string) {
// 	that.Dir = dir
// }

// SetProcStartSecs 设置启动多少秒后没有异常，则表示启动成功
func ProcStartSecs(t int) Option {
	return func(p *ProcessPlus) {
		p.StartSecs = t
	}
}

// func (that *ProcessPlus) SetProcStartSecs(t int) {
// 	that.StartSecs = t
// }

// SetProcExitCodes 设置进程退出的code值列表，该列表中的值表示已知
func ProcExitCodes(codes ...int) Option {
	return func(p *ProcessPlus) {
		p.ExitCodes = codes
	}
}

// func (that *ProcessPlus) SetProcExitCodes(codes ...int) {
// 	that.ExitCodes = codes
// }

// SetProcStartRetries 设置启动失败自动重试次数，默认是3
func ProcStartRetries(rts int) Option {
	return func(p *ProcessPlus) {
		p.StartRetries = rts
	}
}

// func (that *ProcessPlus) SetProcStartRetries(rts int) {
// 	that.StartRetries = rts
// }

// SetProcRestartPause 设置进程重启间隔秒数，默认是0，表示不间隔
func ProcRestartPause(t int) Option {
	return func(p *ProcessPlus) {
		p.RestartPause = t
	}
}

// func (that *ProcessPlus) SetProcRestartPause(t int) {
// 	that.RestartPause = t
// }

// SetProcUser 设置用哪个用户启动进程，默认是父进程的所属用户
func ProcUser(user string) Option {
	return func(p *ProcessPlus) {
		p.User = user
	}
}

// func (that *ProcessPlus) SetProcUser(user string) {
// 	that.User = user
// }

// SetProcPriority 设置进程启动优先级，默认999，值小的优先启动
func ProcPriority(pri int) Option {
	return func(p *ProcessPlus) {
		p.Priority = pri
	}
}

// func (that *ProcessPlus) SetProcPriority(pri int) {
// 	that.Priority = pri
// }

// SetProcStopAsGroup 默认为false,进程被杀死时，是否向这个进程组发送stop信号，包括子进程
func ProcStopAsGroup(sag bool) Option {
	return func(p *ProcessPlus) {
		p.StopAsGroup = sag
	}
}

// func (that *ProcessPlus) SetProcStopAsGroup(sag bool) {
// 	that.StopAsGroup = sag
// }

// SetProcKillAsGroup 默认为false，向进程组发送kill信号，包括子进程
func ProcKillAsGroup(kag bool) Option {
	return func(p *ProcessPlus) {
		p.KillAsGroup = kag
	}
}

// func (that *ProcessPlus) SetProcKillAsGroup(kag bool) {
// 	that.KillAsGroup = kag
// }

// SetProcStopWaitSecs 设置发送结束进程的信号后等待的秒数
func ProcStopWaitSecs(t int) Option {
	return func(p *ProcessPlus) {
		p.StopWaitSecs = t
	}
}

// func (that *ProcessPlus) SetProcStopWaitSecs(t int) {
// 	that.StopWaitSecs = t
// }

// ProcKillWaitSecs 设置强杀进程等待秒数
func ProcKillWaitSecs(t int) Option {
	return func(p *ProcessPlus) {
		p.KillWaitSecs = t
	}
}

// func (that *ProcessPlus) SetProcKillWaitSecs(t int) {
// 	that.KillWaitSecs = t
// }

// SetProcRestartWhenBinaryChanged 当进程的二进制文件有修改，是否需要重启
func ProcRestartWhenBinaryChanged(rwc bool) Option {
	return func(p *ProcessPlus) {
		p.RestartWhenBinaryChanged = rwc
	}
}

// func (that *ProcessPlus) SetProcRestartWhenBinaryChanged(opt bool) {
// 	that.RestartWhenBinaryChanged = opt
// }

// ProcSetExtend 设置扩展参数
func ProcSetExtend(key, value interface{}) Option {
	return func(p *ProcessPlus) {
		p.Extend.Set(key, value)
	}
}

// func (that *ProcessPlus) SetProcSetExtend(key, val interface{}) {
// 	that.Extend.Set(key, val)
// }

// ProcStderrLog 设置stderrlog的存放配置
func ProcStderrLog(file, maxBytes string, backups ...int) Option {
	return func(p *ProcessPlus) {
		p.StderrLogfile = file
		p.StderrLogFileMaxBytes = utils.GetBytes(maxBytes, 50*1024*1024)
		p.StderrLogFileBackups = 10
		if len(backups) > 0 {
			p.StderrLogFileBackups = backups[0]
		}
	}
}

// func (that *ProcessPlus) SetProcStderrLog(file string, maxBytes string, backups ...int) {
// 	that.StderrLogfile = file
// 	that.StderrLogFileMaxBytes = utils.GetBytes(maxBytes, 50*1024*1024)
// 	that.StderrLogFileBackups = 10
// 	if len(backups) > 0 {
// 		that.StderrLogFileBackups = backups[0]
// 	}
// }

// 进程默认配置
func GetDefaultProcSettings() *ProcSettings {
	return &ProcSettings{
		AutoStart:                true,
		StartSecs:                1,
		AutoReStart:              AutoReStartTrue,
		StartRetries:             3,
		RestartPause:             0,
		StopWaitSecs:             10,
		KillWaitSecs:             2,
		Priority:                 999,
		StopAsGroup:              false,
		KillAsGroup:              false,
		RestartWhenBinaryChanged: false,
		Extend:                   gmap.New(true),
		Environment:              gmap.NewStrStrMap(true),
		StdoutLogfile:            "",
		StdoutLogFileMaxBytes:    50 * 1024 * 1024,
		StdoutLogFileBackups:     10,
		RedirectStderr:           false,
		StderrLogfile:            "",
		StderrLogFileMaxBytes:    50 * 1024 * 1024,
		StderrLogFileBackups:     10,
		//User:                     "root",
	}
}
