package processes

import (
	"fmt"
	"os"

	"github.com/moqsien/processes/logger"
	"github.com/moqsien/processes/signals"
)

// Signal 向进程发送信号
// sig: 要发送的信号
// sigChildren: 如果为true，则信号会发送到该进程的子进程
func (that *ProcessPlus) Signal(sig os.Signal, sigChildren bool) error {
	that.Lock.RLock()
	defer that.Lock.RUnlock()

	return that.SendSignal(sig, sigChildren)
}

// 发送多个信号到进程
// sig: 要发送的信号列表
// sigChildren: 如果为true，则信号会发送到该进程的子进程
func (that *ProcessPlus) SendSignals(sigs []string, sigChildren bool) {
	that.Lock.RLock()
	defer that.Lock.RUnlock()

	for _, strSig := range sigs {
		sig := signals.ToSignal(strSig)
		err := that.SendSignal(sig, sigChildren)
		if err != nil {
			logger.Infof("向进程[%s]发送信号[%s]失败,err:%v", that.Name, strSig, err)
		}
	}
}

// sendSignal 向进程发送信号
// sig: 要发送的信号
// sigChildren: 如果为true，则信号会发送到该进程的子进程
func (that *ProcessPlus) SendSignal(sig os.Signal, sigChildren bool) error {
	if that.Cmd != nil && that.Process != nil {
		logger.Infof("发送信号[%s]到进程[%s]", sig, that.Name)
		err := signals.Kill(that.Process, sig, sigChildren)
		return err
	}
	return fmt.Errorf("进程[%s]没有启动", that.Name)
}
