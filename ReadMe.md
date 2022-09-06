### linux下go语言，进程管理
- [x] 提供日志功能
- [x] 提供进程自动重启功能
- [x] 启动失败自动重试 
- [x] 进程启动成功确认(过多少秒之后检查一次，进程仍在运行，则说明成功) 
- [x] 提供进程管理功能
- [x] 进程平滑重启

### 使用方法
```go
manager := NewProcManager()
path := os.Args[0]
name := "test"
// 如果不传path，默认为os.Args[0]
process, _ := manager.NewProcess(name,
                                processes.ProcPath(path),
                                processes.ProcArgs([]string{"go", "get", "xxx"}))
process.StartProc(true)
```
[简单示例](https://github.com/moqsien/processes/examples)

### 设计原理
1、exec.Cmd创建进程，执行外部命令；

2、Process异步起动，可以传入wait参数阻塞父goroutine

3、ProcManager管理正在运行的进程

4、通过向进程发送SIGNAL控制进程的退出

5、通过Clone实现平滑重启

### Thanks To
[DMicro](https://github.com/osgochina/dmicro/)
