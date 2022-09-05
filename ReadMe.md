### 进程管理工具

- [x] 提供日志功能
- [x] 提供进程自动重启功能
- [x] 提供进程管理功能

### 使用方法
```go
manager := NewProcManager()
path := os.Args[0]
name := "test"
process, _ := manager.NewProcess(path, name)
process.SetProcArgs([]string{}) // 设置启动参数
process.SetProcEnvVar("test", "test") // 设置环境变量
process.StartProc(true)
```

### 设计原理
1、exec.Cmd创建进程，执行外部命令；
2、Process用一个goroutine对应一个真实进程，通过goroutine的挂起和唤醒向真实进程发送SIGNAL来控制进程
3、ProcManager管理正在运行的进程

### Thanks To
[DMicro](https://github.com/osgochina/dmicro/)
