package memo_test

import (
	"io/ioutil"
	"log"
	"memo"
	"net/http"
	"sync"
	"testing"
	"time"
)

func httpGetBody(url string) (interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func incomingURLs() <-chan string {
	ch := make(chan string)
	go func() {
		for _, url := range []string{
			"http://golang.org",
			"http://godoc.org",
			"http://play.golang.org",
			"http://golang.io",
			"http://golang.org",
			"http://godoc.org",
			"http://play.golang.org",
			"http://gopl.io",
		} {
			ch <- url
		}
		close(ch)
	}()
	return ch
}

type M interface {
	Get(key string) (interface{}, error)
}

func Sequential(t *testing.T, m M) {
	for url := range incomingURLs() {
		start := time.Now()
		value, err := m.Get(url)
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("%s, %s, %d bytes\n", url, time.Since(start), len(value.([]byte)))
	}
}

func Concurrent(t *testing.T, m M) {
	var n sync.WaitGroup
	for url := range incomingURLs() {
		n.Add(1)
		go func(url string) {
			defer n.Done()
			start := time.Now()
			value, err := m.Get(url)
			if err != nil {
				log.Print(err)
				return
			}
			log.Printf("%s, %s, %d bytes\n", url, time.Since(start), len(value.([]byte)))
		}(url)
	}
	n.Wait()
}

/* 方案一： 有竞态问题 */
func TestSequential(t *testing.T) {
	log.Println("\033[0;31m====> 方案一： 有竞态问题，串行执行\033[0m")
	m := memo.New(httpGetBody)
	Sequential(t, m)
}

func TestConcurrent(t *testing.T) {
	log.Println("\033[0;31m====> 方案一： 有竞态问题，并行执行\033[0m")
	m := memo.New(httpGetBody)
	Concurrent(t, m)
}

/* 方案二： 使用互斥量，但是退化为顺序执行 */
func TestSequential2(t *testing.T) {
	log.Println("\033[0;31m====> 方案二： 使用互斥量，但是退化为顺序执行，串行执行\033[0m")
	m := memo.New2(httpGetBody)
	Sequential(t, m)
}

func TestConcurrent2(t *testing.T) {
	log.Println("\033[0;31m====> 方案二： 使用互斥量，但是退化为顺序执行，并行执行\033[0m")
	m := memo.New2(httpGetBody)
	Concurrent(t, m)
}

/* 方案三：使用互斥量，只将memo的访问放入临界区 */
func TestConcurrent3(t *testing.T) {
	log.Println("\033[0;31m====> 方案三： 使用互斥量，只将memo的访问放入临界区，并行执行\033[0m")
	m := memo.New3(httpGetBody)
	Concurrent(t, m)
}

/* 方案四： 使用互斥量+channel，解决方案二和三的问题 */
func TestConcurrent4(t *testing.T) {
	log.Println("\033[0;31m====> 方案四： 使用互斥量+channel，但是有重复计算的问题，并行执行\033[0m")
	m := memo.New4(httpGetBody)
	Concurrent(t, m)
}

/* 方案五: 使用无锁的channel方案 */
func TestConcurrent5(t *testing.T) {
	log.Println("\033[0;31m====> 方案五： 无锁的channel方案，并行执行\033[0m")
	m := memo.New5(httpGetBody)
	defer m.Close()
	Concurrent(t, m)
}
