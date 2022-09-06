package main

import "github.com/moqsien/processes"

/*
	运行结果示例：

2022-09-06 20:41:21.428 [INFO] 添加进程: test
2022-09-06 20:41:21.428 [INFO] 尝试启动程序[test]
2022-09-06 20:41:21.429 [DEBU] 进程正在运行[test]等待退出
bin
games
include
lib
lib32
lib64
libexec
libx32
local
sbin
share
src
2022-09-06 20:41:21.430 [INFO] 程序[test]已经结束运行，退出码为:exit status 0
*/

func main() {
	manager := processes.NewProcManager()
	process, _ := manager.NewProcess("test",
		processes.ProcPath("/usr/bin/ls"),
		processes.ProcArgs([]string{"/usr"}),
		processes.ProcStdoutLog("/dev/stdout", ""),
	)
	process.StartProc(true)
}
