package main

import (
    `container/list`
    `sync`
)

type Channel struct {
    cond *sync.Cond
    size int
    l *list.List
}

func NewChannel(size int) *Channel{
    return &Channel{
        cond:sync.NewCond(&sync.Mutex{}),
        size:size,
        l:list.New(),
    }
}

func (c *Channel) Send(i int) {
    c.cond.L.Lock()
    defer func() {
        c.cond.Signal()
        c.cond.L.Unlock()
    }()
    // 如果channel buffer满了，阻塞
    for c.l.Len() >= c.size {
        c.cond.Wait()
    }
    c.l.PushBack(i)
}
func (c *Channel) Recv() (i int){
    c.cond.L.Lock()
    for c.l.Len() == 0 {
        c.cond.Wait()
    }
    defer func() {
        c.cond.Signal()
        c.cond.L.Unlock()
    }()
    return c.l.Remove(c.l.Front()).(int)
}


func main() {
    // 创建一个buffer为2的"channel"
    ch := NewChannel(2)
    ch.Send(1)
    ch.Send(2)
    // buffer已满，再发会引发死锁
    // ch.Send(3)
}