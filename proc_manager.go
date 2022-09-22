package processes

import (
	"fmt"
	"os"
	"sync"

	"github.com/gogf/gf/container/gmap"
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
	*gmap.StrAnyMap
}

func NewManager() *Manager {
	return &Manager{
		StrAnyMap: gmap.NewStrAnyMap(),
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
	that.Add(name, p) // 添加进程
	return p, nil
}

// Add 添加进程，重复添加时会覆盖
func (that *Manager) Add(name string, process IProc) {
	// 使用IProc接口作为参数，方便外部对ProcessPlus进行封装
	that.StrAnyMap.Set(name, process)
}

// Search 查找进程
func (that *Manager) SearchProc(name string) (value IProc, found bool) {
	v, found := that.Search(name)
	if found {
		value = v.(IProc)
	}
	return
}

// Remove 从列表移除进程
func (that *Manager) Remove(name string) (value IProc) {
	that.StrAnyMap.Remove(name)
	return
}

// StopAllProcs 停止所有进程
func (that *Manager) StopAllProcs() {
	var wg sync.WaitGroup
	that.Iterator(func(_ string, value interface{}) bool {
		wg.Add(1)
		go func(w *sync.WaitGroup) {
			defer w.Done()
			value.(IProc).StopProc(true)
		}(&wg)
		return true
	})
	wg.Wait()
}

// GetAllProcs 获取所有进程的列表
func (that *Manager) GetAllProcs() []IProc {
	tmpProcList := make([]IProc, 0)
	that.Iterator(func(_ string, value interface{}) bool {
		tmpProcList = append(tmpProcList, value.(IProc))
		return true
	})
	return tmpProcList
}

// GetAllProcsInfo 获取所有进程的信息列表
func (that *Manager) GetAllProcsInfo() ([]*Info, error) {
	AllProcessInfo := make([]*Info, 0)
	that.Iterator(func(_ string, value interface{}) bool {
		procInfo := value.(IProc).GetProcessInfo()
		AllProcessInfo = append(AllProcessInfo, procInfo)
		return true
	})
	return AllProcessInfo, nil
}

// GracefulReload 平滑重启
func (that *Manager) GracefulReload(name string, wait bool) (bool, error) {
	logger.Infof("平滑重启进程[%s]", name)
	p, ok := that.Search(name)
	if !ok {
		return false, fmt.Errorf("没有找到要重启的进程[%s]", name)
	}

	proc := p.(IProc)
	procClone, err := proc.Clone()
	if err != nil {
		return false, err
	}
	procClone.StartProc(wait)
	proc.StopProc(wait)
	that.Add(name, procClone)
	return ok, nil
}
