package processes

import (
	"fmt"
	"os"
	"sync"

	"github.com/gogf/gf/errors/gerror"
	"github.com/moqsien/processes/logger"
)

type IProc interface {
	StartProc(wait bool)
	StopProc(wait bool)
	GetProcessInfo() *Info
	Clone() (*ProcessPlus, error)
}

type Manager struct {
	ProcessList map[string]IProc
	*sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		ProcessList: map[string]IProc{},
		RWMutex:     &sync.RWMutex{},
	}
}

func (that *Manager) NewProcess(name string, options ...Option) (p *ProcessPlus, err error) {
	p = NewProcess(os.Args[0], name)
	if _, found := that.Search(p.Name); found {
		return nil, gerror.Newf("进程[%s]已存在", p.Name)
	}
	p.ProcManager = that
	p.ProcSettings = GetDefaultProcSettings() // 先加载默认配置，再根据options进行修改
	if len(options) > 0 {
		for _, option := range options {
			option(p)
		}
	}
	// that.Add(name, p) // 新进程加入进程管理器中
	return p, nil
}

// Search 查找进程
func (that *Manager) Search(name string) (value IProc, found bool) {
	that.RLock()
	defer that.RUnlock()
	if that.ProcessList != nil {
		value, found = that.ProcessList[name]
	}
	return
}

// Add 添加进程
func (that *Manager) Add(name string, process IProc) {
	that.Lock()
	defer that.Unlock()
	if that.ProcessList == nil {
		that.ProcessList = make(map[string]IProc)
	}
	that.ProcessList[name] = process
}

// Remove 从列表移除进程
func (that *Manager) Remove(name string) (value IProc) {
	that.Lock()
	defer that.Unlock()
	if that.ProcessList != nil {
		var ok bool
		if value, ok = that.ProcessList[name]; ok {
			delete(that.ProcessList, name)
		}
	}
	return
}

// Clear 清空进程列表
func (that *Manager) Clear() {
	that.Lock()
	that.ProcessList = make(map[string]IProc)
	that.Unlock()
}

// Iterator 对进程列表进行迭代
func (that *Manager) Iterate(f func(key string, value IProc) bool) {
	that.RLock()
	defer that.RUnlock()
	for k, v := range that.ProcessList {
		if !f(k, v) {
			break
		}
	}
}

// StopAllProcs 停止所有进程
func (that *Manager) StopAllProcs() {
	var wg sync.WaitGroup
	that.Iterate(func(_ string, proc IProc) bool {
		wg.Add(1)
		go func(w *sync.WaitGroup) {
			defer w.Done()
			proc.StopProc(true)
		}(&wg)
		return true
	})
	wg.Wait()
}

// GetAllProcs 获取所有进程的列表
func (that *Manager) GetAllProcs() []IProc {
	tmpProcList := make([]IProc, 0)
	for _, proc := range that.ProcessList {
		tmpProcList = append(tmpProcList, proc.(*ProcessPlus))
	}
	return tmpProcList
}

// GetAllProcsInfo 获取所有进程的信息列表
func (that *Manager) GetAllProcsInfo() ([]*Info, error) {
	AllProcessInfo := make([]*Info, 0)
	that.Iterate(func(_ string, proc IProc) bool {
		procInfo := proc.GetProcessInfo()
		AllProcessInfo = append(AllProcessInfo, procInfo)
		return true
	})
	return AllProcessInfo, nil
}

// GracefulReload 平滑重启
func (that *Manager) GracefulReload(name string, wait bool) (bool, error) {
	logger.Infof("平滑重启进程[%s]", name)
	proc, ok := that.Search(name)
	if !ok {
		return false, fmt.Errorf("没有找到要重启的进程[%s]", name)
	}
	procClone, err := proc.Clone()
	if err != nil {
		return false, err
	}
	procClone.StartProc(wait)
	proc.StopProc(wait)
	that.Add(name, procClone)
	return true, nil
}
