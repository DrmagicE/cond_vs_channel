package main

import (
	"container/list"
	"fmt"
	"sync"
)

type Channel struct {
	cond *sync.Cond
	size int
	// 表示Channel是否被关闭
	close      bool
	caseNumber int
	l          *list.List
}

func NewChannel(size int) *Channel {
	return &Channel{
		cond: sync.NewCond(&sync.Mutex{}),
		size: size,
		l:    list.New(),
	}
}

func (c *Channel) Send(i int) {
	c.cond.L.Lock()
	defer func() {
		c.cond.Signal()
		c.cond.L.Unlock()
	}()
	// 如果channel buffer满了，阻塞
	for c.l.Len() >= c.size && !c.close {
		c.cond.Wait()
	}
	// 向已关闭的"Channel"写入数据会引发panic
	if c.close {
		panic("send on closed channel")
	}
	c.l.PushBack(i)
}

// 为了能在"Channel"关闭的时候，读出所有未读的数据，将返回值修改成数组
func (c *Channel) Recv() (i []int) {
	c.cond.L.Lock()
	// 如果"Channel"被关闭了，则不需要阻塞等待
	for c.l.Len() == 0 && !c.close {
		c.cond.Wait()
	}
	defer func() {
		c.cond.Signal()
		c.cond.L.Unlock()
	}()
	if c.close {
		// "Channel"关闭，返回所有未读的数据
		for e := c.l.Front(); e != nil; {
			curent := e
			e = e.Next()
			i = append(i, c.l.Remove(curent).(int))
		}
		return i
	}
	return []int{c.l.Remove(c.l.Front()).(int)}
}

func (c *Channel) Close() {
	c.cond.L.Lock()
	defer func() {
		c.cond.Signal()
		c.cond.L.Unlock()
	}()
	c.close = true
}

func (c *Channel) Open() {
	c.cond.L.Lock()
	defer func() {
		c.cond.Signal()
		c.cond.L.Unlock()
	}()
	c.close = false
}

func main() {
	// 创建一个buffer为2的"channel"
	ch := NewChannel(2)
	ch.Send(1)
	ch.Send(2)
	// 关闭channel
	ch.Close()
	fmt.Println(ch.Recv()) // [1,2]
	// 不会阻塞，返回空
	fmt.Println(ch.Recv()) // []
	// 向已关闭的"Channel"写入数据，panic
	ch.Send(1)
}
