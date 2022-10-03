package memo

import (
	// "log"
	"sync"
)

/* 方案一: 有竞态问题 */
type result struct {
	value interface{}
	err   error
}

type Func func(key string) (interface{}, error)

type Memo struct {
	f     Func
	cache map[string]result
}

func New(f Func) *Memo {
	return &Memo{f: f, cache: make(map[string]result)}
}

func (memo *Memo) Get(key string) (interface{}, error) {
	res, ok := memo.cache[key]
	if !ok {
		res.value, res.err = memo.f(key)
		memo.cache[key] = res
	}
	return res.value, res.err
}

/* 方案二: 使用互斥量同步 */
type Memo2 struct {
	f     Func
	mu    sync.Mutex
	cache map[string]result
}

func New2(f Func) *Memo2 {
	return &Memo2{f: f, cache: make(map[string]result)}
}

/* memo.f(key)的操作也在临界区中执行
 * 导致整体退化为一个串行操作
 */
func (memo *Memo2) Get(key string) (interface{}, error) {
	memo.mu.Lock()
	defer memo.mu.Unlock()
	res, ok := memo.cache[key]
	if !ok {
		res.value, res.err = memo.f(key)
		memo.cache[key] = res
	}
	return res.value, res.err
}

/* 方案三: 使用互斥量，较之方案二，
 * 减少临界区范围，使得memo.f(key)可以并行
 * 只将对Memo2变量的访问放入临界区
 * 但是有可能多个goroutine同时执行memo.f(key)
 * 虽然最终只有一个结果会被写入缓存
 */
type Memo3 Memo2

func New3(f Func) *Memo3 {
	return &Memo3{f: f, cache: make(map[string]result)}
}

func (memo *Memo3) Get(key string) (interface{}, error) {
	memo.mu.Lock()
	res, ok := memo.cache[key]
	memo.mu.Unlock()
	if !ok {
		res.value, res.err = memo.f(key)
		memo.mu.Lock()
		memo.cache[key] = res
		memo.mu.Unlock()
	}
	return res.value, res.err
}

/* 方案四: 使用channel + 互斥量
 * 方案三的多次写入的问题是由于两次加锁导致的
 * 而归并为一次加锁就是方案二，退化为串行执行
 * 本质是memo.f(key)是一个耗时操作，串行执行效率低
 * 如果能有某种方法，快速设计标志，表示已经有goroutine执行memo.f(key)，
 * 其它goroutine不要重复执行，等待对应goroutine执行完成
 * 问题简化为同步另一goroutine，使用channel可以解决该问题
 * 	- channel读取阻塞表明数据尚未缓存
 * 	- 利用读取关闭的channel立刻返回元素类型的零值这一特性，来表明数据已经缓存
 */

type entry struct {
	res   result
	ready chan struct{}
}

type Memo4 struct {
	f     Func
	mu    sync.Mutex
	cache map[string]*entry
}

func New4(f Func) *Memo4 {
	return &Memo4{f: f, cache: make(map[string]*entry)}
}

func (memo *Memo4) Get(key string) (interface{}, error) {
	memo.mu.Lock()
	e := memo.cache[key] // 不存在返回零值，指针的零值即nil
	if e == nil {
		e = &entry{ready: make(chan struct{})}
		memo.cache[key] = e
		memo.mu.Unlock()
		e.res.value, e.res.err = memo.f(key)
		close(e.ready)
	} else {
		memo.mu.Unlock()
		<-e.ready
	}
	return e.res.value, e.res.err
}

/* 方案五： 使用channel，无锁方案
 * 无锁意味着只有一个goroutine能访问用于缓存的变量，称之为goroutine X
 * 执行Get的goroutine需要和真正访问的goroutine通过channel通信
 * 首先需要将请求发送给goroutine X, "request_chan <- request"
 * 然后等待返回结果 "<- response_chan"
 * 每个goroutine需要等待各自的结果，显然每个goroutine用于返回reponse的channel都是自有的
 * 该response channel需要传递给goroutine X，用于X发送结果数据
 * 即request channel中需要携带response channel和key信息
 */

type request struct {
	key      string
	response chan<- result
}

type Memo5 struct {
	requests chan request
}

func New5(f Func) (m *Memo5) {
	m = &Memo5{requests: make(chan request)}
	go m.server(f)
	return m
}

func (memo *Memo5) Get(key string) (interface{}, error) {
	response := make(chan result)
	memo.requests <- request{key: key, response: response}
	res := <-response
	return res.value, res.err
}

func (memo *Memo5) server(f Func) {
	cache := make(map[string]*entry)
	for req := range memo.requests {
		if cache[req.key] == nil {
			e := &entry{ready: make(chan struct{})}
			cache[req.key] = e
			// 必须使用goroutine，否则退化为串行执行f函数
			go e.call(f, req.key)
		}
		// 必须使用goroutine，否则阻塞等待，本质也将退化为串行执行函数
		go cache[req.key].deliver(req.response)
	}
}

func (memo *Memo5) Close() { close(memo.requests) }

func (e *entry) call(f Func, key string) {
	e.res.value, e.res.err = f(key)
	close(e.ready)
}

func (e *entry) deliver(response chan<- result) {
	<-e.ready
	response <- e.res
}
