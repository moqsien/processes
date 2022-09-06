package processes

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gogf/gf/os/genv"
	"github.com/moqsien/processes/logger"
	"github.com/moqsien/processes/proclog"
	"github.com/moqsien/processes/signals"
)

type ProcessPlus struct {
	*exec.Cmd
	*ProcSettings
	ProcManager *ProcManager // 进程管理器
	Name        string       // 进程名称
	State       ProcState    // 进程的当前状态
	Starting    bool         // 正在启动的时候，该值为true
	StopByUser  bool         // 用户主动关闭的时候，该值为true
	RetryTimes  *int32       // 启动重试的次数
	StartTime   time.Time    // 启动时间
	StopTime    time.Time    // 停止时间

	Lock      sync.RWMutex
	Stdin     io.WriteCloser
	StdoutLog proclog.Logger
	StderrLog proclog.Logger
}

// NewProcess 创建进程: path, 可执行文件绝对路径；name, 进程名称
func NewProcess(path, name string) (p *ProcessPlus) {
	p = &ProcessPlus{}
	p.Cmd = exec.Command(path)
	if len(p.Args) == 0 {
		p.Args = []string{path}
	}
	p.Name = name
	// 父进程退出，则它生成的子进程也全部退出
	p.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGKILL,
	}
	p.RetryTimes = new(int32)
	return
}

func (that *ProcessPlus) Init() (err error) {
	// 设置进程运行的环境变量
	if that.Environment.Size() > 0 {
		_ = genv.SetMap(that.Environment.Map())
	}
	that.Env = genv.All()

	// 设置程序运行时用户
	if that.SetUser() != nil {
		err = fmt.Errorf("设置程序运行时用户[%s]失败", that.User)
		return
	}

	// 设置进程的运行日志存放文件
	that.StdoutLog = that.CreateStdoutLogger()
	that.Stdout = that.StdoutLog
	if that.RedirectStderr {
		that.StderrLog = that.StdoutLog
	} else {
		that.StderrLog = that.CreateStderrLogger()
	}
	that.Stderr = that.StderrLog

	// 程序的标准输入
	that.Stdin, _ = that.StdinPipe()
	return
}

// Clone 克隆进程
func (that *ProcessPlus) Clone() (*ProcessPlus, error) {
	proc := NewProcess(that.Path, that.Name)
	proc.ProcManager = that.ProcManager

	proc.ProcSettings = &(*that.ProcSettings)
	proc.ProcSettings.Environment = proc.ProcSettings.Environment.Clone()
	proc.ProcSettings.Extend = proc.ProcSettings.Extend.Clone()

	proc.StartTime = time.Unix(0, 0)
	proc.StopTime = time.Unix(0, 0)
	proc.State = Stopped
	proc.Starting = false
	proc.StopByUser = false
	proc.RetryTimes = new(int32)
	proc.Init()
	return proc, nil
}

// 监控进程是否正在运行中
func (that *ProcessPlus) MonitorProgramIsRunning(endTime time.Time, monitorExited *int32, programExited *int32) {
	for time.Now().Before(endTime) && atomic.LoadInt32(programExited) == 0 {
		// 每100毫秒空转一次，直到时间到达或者进程退出
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	defer atomic.StoreInt32(monitorExited, 1) // 修改监控goroutine的退出状态
	if atomic.LoadInt32(programExited) == 0 && that.State == Starting {
		// 进程启动成功，则修改State
		logger.Infof("进程[%s]启动成功", that.Name)
		that.Lock.Lock()
		that.State = Running
		that.Lock.Unlock()
	}
}

// 设置程序启动失败状态
func (that *ProcessPlus) FailToStartProgram(reason string, finishCb func()) {
	logger.Errorf("程序[%s]启动失败，失败原因：%s ", that.Name, reason)
	that.State = Fatal
	finishCb()
}

// 获取配置的退出code值列表
func (that *ProcessPlus) GetExitCodes() []int {
	strExitCodes := that.ExitCodes
	if len(that.ExitCodes) > 0 {
		return strExitCodes
	}
	return []int{0, 2}
}

// 进程的退出code值是否在设置中的codes列表中
func (that *ProcessPlus) InExitCodes(exitCode int) bool {
	for _, code := range that.GetExitCodes() {
		if code == exitCode {
			return true
		}
	}
	return false
}

// 获取进程的退出code值
func (that *ProcessPlus) GetExitCode() (int, error) {
	if that.ProcessState == nil {
		return -1, fmt.Errorf("no exit code")
	}
	if status, ok := that.ProcessState.Sys().(syscall.WaitStatus); ok {
		return status.ExitStatus(), nil
	}

	return -1, fmt.Errorf("no exit code")

}

// 判断进程是否在运行
func (that *ProcessPlus) IsRunning() bool {
	if that.Cmd != nil && that.Process != nil {
		if runtime.GOOS == "windows" {
			proc, err := os.FindProcess(that.Process.Pid)
			return proc != nil && err == nil
		}
		return that.Process.Signal(syscall.Signal(0)) == nil
	}
	return false
}

// 判断进程是否需要自动重启
func (that *ProcessPlus) IsAutoRestart() bool {
	autoRestart := that.AutoReStart

	if autoRestart == AutoReStartFalse {
		return false
	} else if autoRestart == AutoReStartTrue {
		return true
	} else {
		that.Lock.RLock()
		defer that.Lock.RUnlock()
		if that.Cmd != nil && that.ProcessState != nil {
			exitCode, err := that.GetExitCode()
			// 如果自动重启设置为unexpected，则表示，在配置中已明确的退出code不需要重启，
			// 不在预设的配置中的退出code则需要重启
			return err == nil && !that.InExitCodes(exitCode)
		}
	}
	return false
}

// 阻塞等待进程运行结束
func (that *ProcessPlus) WaitForExit(_ int64) {
	_ = that.Wait()
	// 进程退出后执行
	if that.ProcessState != nil {
		logger.Infof("程序[%s]已经结束运行，退出码为:%v", that.Name, that.ProcessState)
	} else {
		logger.Infof("程序[%s]已经结束运行", that.Name)
	}
	that.Lock.Lock()
	defer that.Lock.Unlock()
	that.StopTime = time.Now()

	// 关闭标准输出
	if that.StdoutLog != nil {
		_ = that.StdoutLog.Close()
	}
	if that.StderrLog != nil {
		_ = that.StderrLog.Close()
	}
}

// Stop 主动停止进程
func (that *ProcessPlus) StopProc(wait bool) {

	that.Lock.Lock()
	that.StopByUser = true
	isRunning := that.IsRunning()
	that.Lock.Unlock()

	if !isRunning {
		logger.Infof("程序[%s]未运行", that.Name)
		return
	}
	logger.Infof("正在停止程序[%s]", that.Name)

	// 获取程序的正常退出信号
	sigs := that.StopSignal
	// 发送信号后的等待秒数
	waitSecond := time.Duration(that.StopWaitSecs) * time.Second
	// 发送强制结束信号后的等待秒数
	killWaitSecond := time.Duration(that.KillWaitSecs) * time.Second
	// 是否同时停止进程组
	stopAsGroup := that.StopAsGroup
	// 是否强制杀死进程组
	killAsGroup := that.KillAsGroup
	if stopAsGroup && !killAsGroup {
		logger.Error("不能够同时设置 stopAsGroup=true 和 killAsGroup=false")
	}
	var stopped int32 = 0

	go func() {
		for i := 0; i < len(sigs) && atomic.LoadInt32(&stopped) == 0; i++ {
			// 获取需要发送的信号
			sig := signals.ToSignal(sigs[i])
			logger.Infof("发送结束进程信号[%s]给进程[%s]", that.Name, sigs[i])
			// 发送结束进程信号给程序，发送信号后，进程正常结束，则RunProc的cmd.Wait()会继续向下执行，并修改进程的State
			_ = that.Signal(sig, stopAsGroup)
			endTime := time.Now().Add(waitSecond)
			//等待指定的时候后，判断当前进程是否还在存
			for endTime.After(time.Now()) {
				// 如果进程成功结束，则State一般是修改为Exited状态
				if (that.State & Exist) == 0 {
					atomic.StoreInt32(&stopped, 1)
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		// 如果发送了设置的信号后，进程还未停止，则需要强制结束该进程
		if atomic.LoadInt32(&stopped) == 0 {
			logger.Infof("强制结束程序[%s]", that.Name)
			_ = that.Signal(syscall.SIGKILL, killAsGroup)
			killEndTime := time.Now().Add(killWaitSecond)
			for killEndTime.After(time.Now()) {
				//如果进程结束成功
				if (that.State & Exist) == 0 {
					atomic.StoreInt32(&stopped, 1)
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			//无论如何，发送了强杀信号后，默认认为它强杀成功
			atomic.StoreInt32(&stopped, 1)
		}
	}()

	// 是否阻塞等待进程结束
	if wait {
		for atomic.LoadInt32(&stopped) == 0 {
			time.Sleep(1 * time.Second)
		}
	}
}

// 运行进程，finishCb是进程创建过程结束之后的回调，用于解除父goroutine的阻塞
func (that *ProcessPlus) RunProc(finishCb func()) {
	that.Lock.Lock()
	defer that.Lock.Unlock()

	// 判断进程是否正在运行
	if that.IsRunning() {
		logger.Infof("不能启动进程[%s],因为它正在运行中...", that.Name)
		finishCb()
		return
	}

	that.StartTime = time.Now()
	atomic.StoreInt32(that.RetryTimes, 0)

	//指定启动多少秒后没有异常退出，则表示启动成功
	startSecs := that.StartSecs
	// 进程重启间隔秒数，默认是0，表示不间隔
	restartPause := that.RestartPause

	var once sync.Once
	finishCbWrapper := func() {
		once.Do(finishCb)
	}

	// 进程被用户结束
	for !that.StopByUser { // 直到用户结束进程或者成功启动进程命令

		//如果进程启动失败，需要重试，则需要判断配置，重试启动是否需要间隔制定时间
		if restartPause > 0 && atomic.LoadInt32(that.RetryTimes) != 0 {
			logger.Infof("不能立刻重启程序[%s],需要等待%d秒", that.Name, restartPause)
			time.Sleep(time.Duration(restartPause) * time.Second)
		}
		// 程序指定结束时间，如果在该时间内未退出，则表示进程启动成功
		endTime := time.Now().Add(time.Duration(startSecs) * time.Second)
		//更新进程状态
		that.State = Starting

		// 启动次数+1
		atomic.AddInt32(that.RetryTimes, 1)

		// 进程初始化
		err := that.Init()
		if err != nil {
			that.FailToStartProgram(fmt.Sprintf("不能创建进程,err:%v", err), finishCbWrapper)
			break
		}

		// 启动程序, Cmd.Start()
		err = that.Start()
		if err != nil {
			// 重试次数已经大于设置中的最大重试次数
			if atomic.LoadInt32(that.RetryTimes) >= int32(that.StartRetries) {
				that.FailToStartProgram(fmt.Sprintf("error:%v", err), finishCbWrapper)
				break
			} else {
				// 启动失败，再次重试
				logger.Infof("程序[%s]启动失败,再次重试,error:%v", that.Name, err)
				that.State = Suspend
				continue
			}
		}

		//设置标准输出日志的pid
		if that.StdoutLog != nil {
			that.StdoutLog.SetPid(that.Pid())
		}
		// 设置标准错误输出日志的pid
		if that.StderrLog != nil {
			that.StderrLog.SetPid(that.Pid())
		}

		monitorExited := int32(0)
		programExited := int32(0)
		// 如果未设置启动监视时长，则表示cmd.start成功就算该程序启动成功
		if startSecs <= 0 {
			logger.Infof("程序[%s]启动成功", that.Name)
			that.State = Running
			go finishCbWrapper()
		} else {
			go func() { // 异步监控进程是否成功运行
				// 启动一段时间后，如果没有退出，就认为启动成功，修改State为Running
				that.MonitorProgramIsRunning(endTime, &monitorExited, &programExited)
				// 进程成功启动，父goroutine解除阻塞
				finishCbWrapper()
			}()
		}
		logger.Debugf("进程正在运行[%s]等待退出", that.Name)
		that.Lock.Unlock() // 先解锁，因为MonitorProgramIsRunning和WaitForExit中需要加解锁

		that.WaitForExit(int64(startSecs))          // 阻塞等待进程退出(执行完毕或者收到退出Signal)后，会关闭标准输出，并修改StopTime
		atomic.StoreInt32(&programExited, 1)        // 进程已经退出，修改进程退出标记，主要是当监控goroutine还在运行时，将其快速结束
		for atomic.LoadInt32(&monitorExited) == 0 { // 等待监控协程退出
			time.Sleep(time.Duration(10) * time.Millisecond)
		}

		that.Lock.Lock() // 加锁，用于对后续状态修改的保护，解锁在defer中进行

		// 如果此时的State为Running，则为进程正常运行并退出
		if that.State == Running {
			that.State = Exited
			logger.Infof("程序[%s]已经结束", that.Name)
			break
		} else { // 如果此时的State为Starting，则说明进程在监控时间内挂掉了(如调用ProcStop或者进程运行出错)，需要重试
			that.State = Suspend
		}

		// 如果重试次数已经超过了设置的最大重试次数
		if atomic.LoadInt32(that.RetryTimes) >= int32(that.StartRetries) {
			that.FailToStartProgram(fmt.Sprintf("不能启动程序[%s],因为已经超出了它的最大重试值:%d", that.Name, that.StartRetries), finishCbWrapper)
			break
		}
	}
}

// Start 启动进程，wait表示阻塞等待进程成功启动
func (that *ProcessPlus) StartProc(wait bool) {
	logger.Infof("尝试启动程序[%s]", that.Name)

	that.Lock.Lock()
	if that.Starting {
		logger.Infof("不成重复启动该进程[%s],因为该进程已经启动！", that.Name)
		that.Lock.Unlock()
		return
	}
	that.Starting = true
	that.StopByUser = false
	that.Lock.Unlock()

	var runCond *sync.Cond
	if wait {
		runCond = sync.NewCond(&sync.Mutex{})
		runCond.L.Lock()
	}

	// 异步起动进程
	go func() {
		for { // 一直尝试运行进程命令，直到成功或者超过最大重试次数
			that.RunProc(func() {
				if wait { // 阻塞等待进程的创建过程
					runCond.L.Lock()
					runCond.Signal() // 进程创建成功或者失败之后，发送信号解除父goroutine的阻塞
					runCond.L.Unlock()
				}
			})

			// 只有用户调用了ProcStop 或者 正常运行结束才会走到下面的逻辑，否则直接跳到runCond.Wait()
			if that.StopByUser {
				logger.Infof("用户主动结束了该程序[%s]，不要再次启动", that.Name)
				break
			}

			// 判断进程是否需要自动重启，一些一次性运行的任务进程，不需要重启
			if !that.IsAutoRestart() {
				logger.Infof("不要自动重启进程[%s],因为该进程设置了不需要自动重启", that.Name)
				break
			}

			// 如果上一次进程启动失败，并且启动时间少于2秒，则需要暂停一会，避免死循环，耗干资源，一般不会走到这里
			if time.Now().Unix()-that.StartTime.Unix() < 2 {
				time.Sleep(3 * time.Second)
			}

			// 对于隔一段时间运行一次的进程，可以重启
			logger.Infof("因为该进程设置了自动重启,自动重启进程[%s],", that.Name)
		}
		that.Lock.Lock()
		that.Starting = false
		that.Lock.Unlock()
	}()

	if wait {
		runCond.Wait() // runCond.L.Unlock()，然后挂起goroutine并等待runCond.Signal()触发
		runCond.L.Unlock()
	}
}
