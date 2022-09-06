package processes

/*=======================
  自动重启模式:
    进程退出后，自动重启的规则
*/

type AutoReStart string

const (
	AutoReStartUnexpected AutoReStart = "unexpected" // 默认为unexpected，表示当进程被意外杀死后才重启
	AutoReStartTrue       AutoReStart = "true"       // 总是自动重启
	AutoReStartFalse      AutoReStart = "false"      // 关闭自动重启功能
)

/*
=======================
进程模式：

	1、多进程
	2、单进程
*/
type ProcessMode int

const (
	SingleProcess     ProcessMode = 0                     // Single 单进程模式
	MultiProcess      ProcessMode = 1                     // Multi 多进程模式
	RemarkOfMasterEnv string      = "MasterProcessRemark" // Remark 多进程模式下，用于标记主进程的环境变量的名称
)

/*
=======================
进程状态：

	描述进程状态的类型
*/
type ProcState int

const (
	Stopped  ProcState = 1   // Stopped 已停止
	Starting ProcState = 2   // Starting 启动中
	Running  ProcState = 4   // Running 运行中
	Suspend  ProcState = 8   // Suspend 已挂起
	Stopping ProcState = 16  // Stopping 停止中
	Exited   ProcState = 32  // Exited 已退出
	Fatal    ProcState = 64  // Fatal 启动失败
	Unknown  ProcState = 128 // Unknown 未知状态
	Failure  ProcState = Stopped | Fatal | Unknown | Exited | Suspend
	Exist    ProcState = Running | Starting | Stopping
)

func (ps *ProcState) ToString() string {
	switch *ps {
	case Stopped:
		return "Stopped"
	case Starting:
		return "Starting"
	case Running:
		return "Running"
	case Suspend:
		return "Suspend"
	case Stopping:
		return "Stopping"
	case Exited:
		return "Exited"
	case Fatal:
		return "Fatal"
	default:
		return "Unknown"
	}
}
