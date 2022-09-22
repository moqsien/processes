package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gogf/gf/container/gtree"
	"github.com/gogf/gf/util/gutil"
	"github.com/moqsien/processes"
)

/*
	运行结果示例
*/
var manager = processes.NewManager()

var ch = make(chan interface{}, 6)

var tree = gtree.NewRedBlackTree(gutil.ComparatorString, true)

func test(key interface{}, _ interface{}) bool {
	name, _ := key.(string)
	process, _ := manager.NewProcess(name,
		processes.ProcPath(os.Args[0]),
		processes.ProcArgs([]string{name}),
		processes.ProcStdoutLog("/dev/stdout", ""),
	)
	process.StartProc(true)
	manager.Add(name, process)
	return true
}

func t(key string) {
	for {
		fmt.Println(">>>process: ", key)
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		tree.Set("a", 1)
		tree.Set("b", 2)
		tree.Set("c", 3)
		tree.Set("d", 4)
		tree.IteratorAsc(test)
	} else if len(args) == 1 {
		fmt.Println(args)
		t(args[0])
	}
	p, found := manager.SearchProc("a")
	if found {
		fmt.Println(p.GetProcessInfo())
	}
}
