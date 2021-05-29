package main

import (
	"container/list"
	"fmt"
	"sync"
)

// 模拟channel
type Channel struct {
	// 条件变量
	cond *sync.Cond
	// l 用于保存channel中的内容
	l *list.List
}

// NewChannel 初始化Channel
func NewChannel() *Channel {
	return &Channel{
		cond: sync.NewCond(&sync.Mutex{}),
		l:    list.New(),
	}
}

// Send 向Channel中发送数据
func (c *Channel) Send(i int) {
	c.cond.L.Lock()
	defer func() {
		c.cond.Signal()
		c.cond.L.Unlock()
	}()
	c.l.PushBack(i)
}

// Recv 接收数据，如果Channel中没有内容，则阻塞等待。
func (c *Channel) Recv() (i int) {
	// Send 和 Recv都需要访问c.l，要加锁
	c.cond.L.Lock()
	for c.l.Len() == 0 {
		// 如果channel中还没有数据，那么Wait等待Signal或Broadcast的通知。
		// 在Wait()方法中，会包含Unlock的逻辑，当收到信号后，会重新Lock。
		c.cond.Wait()
		// 这里一定要用for而不能用if，因为在Wait方法中，收到信号和重新加锁之并不是原子操作。
		// 其他goroutine是有可能在重新加锁之前对c.l其进行修改的（那么len就可能不是0了），因此还要再进入for判断一次。
	}
	defer func() {
		c.cond.Signal()
		c.cond.L.Unlock()
	}()
	return c.l.Remove(c.l.Front()).(int)
}

func main() {
	ch := NewChannel()
	for i := 0; i < 10; i++ {
		go func(i int) {
			ch.Send(i)
		}(i)
	}
	for i := 0; i < 10; i++ {
		fmt.Println(ch.Recv())
	}
	// 死锁
	//fmt.Println(ch.Recv())
}
