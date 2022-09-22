package main

import (
	"fmt"
	"os"
	"runtime"
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
	return true
}

func t(key string) {
	for {
		fmt.Println(">>>process: ", key)
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func test2(key interface{}, value interface{}) bool {
	fmt.Println("hello", key, value)
	i, _ := value.(int64)
	time.Sleep(time.Duration(i) * time.Second)
	name, _ := key.(string)
	t(name)
	return true
}

func test3(key, value interface{}) bool {
	fmt.Println("+++", key, value)
	ch <- key
	k, _ := key.(string)
	if k == "d" {
		close(ch)
	}
	return true
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
	fmt.Println(runtime.NumGoroutine())
	// tree := gtree.NewAVLTree(gutil.ComparatorInt)
	// for i := 0; i < 10; i++ {
	// 	tree.Set(i, i*10)
	// }
	// // 打印树形
	// tree.Print()
	// // 前序遍历
	// fmt.Println("ASC:")
	// tree.IteratorAsc(func(key, value interface{}) bool {
	// 	fmt.Println(key, value)
	// 	return true
	// })
	// // 后续遍历
	// fmt.Println("DESC:")
	// tree.IteratorDesc(func(key, value interface{}) bool {
	// 	fmt.Println(key, value)
	// 	return true
	// })
}
