package processes

import (
	"fmt"
	"sync"

	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/errors/gerror"
	"github.com/moqsien/processes/logger"
)

type ProcManager struct {
	Container *gmap.StrAnyMap // 存放ProcessPlus进程对象的容器
}

func NewProcManager() *ProcManager {
	return &ProcManager{
		Container: &gmap.StrAnyMap{},
	}
}

// NewProcess 创建新进程：path，可执行文件路径，一般os.Args[0]获取当前go程序可执行文件路径；name，进程名称
func (that *ProcManager) NewProcess(path, name string, options ...Option) (p *ProcessPlus, err error) {
	p = NewProcess(path, name)
	if _, found := that.Container.Search(p.Name); found {
		return nil, gerror.Newf("进程[%s]已存在", p.Name)
	}
	p.ProcManager = that
	p.ProcSettings = GetDefaultProcSettings() // 先加载默认配置，再根据options进行修改
	if len(options) > 0 {
		for _, option := range options {
			option(p)
		}
	}
	that.Add(name, p) // 新进程加入进程管理器中
	return p, nil
}

// Add 添加进程到Manager
func (that *ProcManager) Add(name string, proc *ProcessPlus) {
	that.Container.Set(name, proc)
	logger.Info("添加进程:", name)
}

// Remove 从Manager移除进程
func (that *ProcManager) Remove(name string) *ProcessPlus {
	proc := that.Container.Remove(name)
	if proc == nil {
		return nil
	}
	logger.Info("remove process:", name)
	return proc.(*ProcessPlus)
}

// Clear 清空容器
func (that *ProcManager) Clear() {
	that.Container.Clear()
}

// ForEachProcess 迭代进程列表
func (that *ProcManager) ForEachProcess(procFunc func(p *ProcessPlus)) {
	that.Container.Iterator(func(_ string, v interface{}) bool {
		procFunc(v.(*ProcessPlus))
		return true
	})
}

// StopAllProcesses 关闭所有进程
func (that *ProcManager) StopAllProcesses() {
	var wg sync.WaitGroup

	that.ForEachProcess(func(proc *ProcessPlus) {
		wg.Add(1)

		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			proc.StopProc(true)
		}(&wg)
	})
	wg.Wait()
}

// 获取所有进程列表
func (that *ProcManager) getAllProcess() []*ProcessPlus {
	tmpProcList := make([]*ProcessPlus, 0)
	for _, proc := range that.Container.Map() {
		tmpProcList = append(tmpProcList, proc.(*ProcessPlus))
	}
	return tmpProcList
}

// Find 根据进程名查询进程
func (that *ProcManager) Find(name string) *ProcessPlus {
	proc, ok := that.Container.Search(name)
	if ok {
		return proc.(*ProcessPlus)
	}
	return nil
}

// GetAllProcessInfo 获取所有进程信息
func (that *ProcManager) GetAllProcessInfo() ([]*Info, error) {
	AllProcessInfo := make([]*Info, 0)
	that.ForEachProcess(func(proc *ProcessPlus) {
		procInfo := proc.GetProcessInfo()
		AllProcessInfo = append(AllProcessInfo, procInfo)
	})
	return AllProcessInfo, nil
}

// GracefulReload 平滑重启进程
func (that *ProcManager) GracefulReload(name string, wait bool) (bool, error) {
	logger.Infof("平滑重启进程[%s]", name)
	proc := that.Find(name)
	if proc == nil {
		return false, fmt.Errorf("没有找到要重启的进程[%s]", name)
	}
	procClone, err := proc.Clone()
	if err != nil {
		return false, err
	}
	procClone.StartProc(wait)
	proc.StopProc(wait)
	that.Container.Set(name, procClone)
	return true, nil
}
