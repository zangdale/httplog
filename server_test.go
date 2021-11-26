package httplog

import (
	"bytes"
	"container/heap"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// 测试简单队列 heap
func TestQueue(t *testing.T) {
	// 堆测试
	h := &eventByteLengthQueue{2, 1, 5}
	heap.Init(h)
	heap.Push(h, 3)
	fmt.Printf("minimum: %d\n", (*h)[0])
	for h.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(h))
	}
}

func TestPushUseByte(t *testing.T) {
	a := make(chan int)
	go func() {
		for i := 0; i < 100; i++ {
			events.Publish([]byte(fmt.Sprintf(`{"qwertyuiopa{sdfghj}klzxcvbnm":"%s"}`, uuid.NewString())))
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			fmt.Println()
			bu := bytes.Buffer{}
			bu.WriteByte('[')
			bu.Write(events.GetAllEvents())
			bu.WriteByte(']')
			log.Info().RawJSON("body", bu.Bytes()).Send()
			time.Sleep(time.Millisecond * 2)
		}
	}()

	<-a
}

func TestByteSlice(t *testing.T) {
	a := []byte("123456")
	n := len(a)
	fmt.Println("len", n)
	fmt.Println("n-1", a[n-1:])
	fmt.Println("n", a[n:])
	fmt.Println("n+1", a[n+1:])

	/*
		len 6
		n-1 [54]
		n []
		runtime error: slice bounds out of range [7:6] [recovered]
			panic: runtime error: slice bounds out of range [7:6]
	*/
}

func BenchmarkPushUseByte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		events.Publish([]byte(fmt.Sprintf("{qqqqqqqqqqqqqqqqqqqqqqqqq{%d}111111111111111111111111111}", i)))
	}
	/*
		goos: windows
		goarch: amd64
		pkg: github.com/zangdale/httplog
		cpu: Intel(R) Core(TM) i5-8250U CPU @ 1.60GHz
		BenchmarkPush
		BenchmarkPush-8
		 2909349	       428.9 ns/op	     434 B/op	       4 allocs/op
		PASS
	*/
}

func BenchmarkPush_UseByte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		events.Publish([]byte("{qqqqqqqqqqqqqqqqqqqqqqqqq{8888888888888888}111111111111111111111111111}"))
	}
	/*
		goos: windows
		goarch: amd64
		pkg: github.com/zangdale/httplog
		cpu: Intel(R) Core(TM) i5-8250U CPU @ 1.60GHz
		BenchmarkPush_
		BenchmarkPush_-8
		 5822178	       205.6 ns/op	     369 B/op	       2 allocs/op
		PASS
	*/
}

// todo bytes.Buffer
