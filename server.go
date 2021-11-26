package httplog

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// 学习使用 堆实现的简单队列 使用 container/heap
// 也可使用 container/list 实现
// 最简单就是直接实现 Push 和Pop 既可以
// An eventByteLengthQueue is a min-queue of ints.
type eventByteLengthQueue []int

func (h eventByteLengthQueue) Len() int           { return len(h) }
func (h eventByteLengthQueue) Less(i, j int) bool { return false }
func (h eventByteLengthQueue) Swap(i, j int)      {}

// Push 往后添加一个数据
func (h *eventByteLengthQueue) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(int))
}

// Pop 拿出前面的一个数据
func (h *eventByteLengthQueue) Pop() interface{} {
	old := *h
	n := len(old)
	if n == 0 {
		return 0
	}

	x := old[0]
	if n == 1 {
		*h = old[1:]
	} else {
		*h = old[1 : n-1]
	}

	return x
}

/////// 正式代码 ///////////////

type eventLengthQueue []int

// push 往后添加一个数据
func (h *eventLengthQueue) push(x int) {
	*h = append(*h, x)
}

// pop 拿出前面的一个数据
func (h *eventLengthQueue) pop() int {
	old := *h
	n := len(old)
	if n == 0 {
		return 0
	}

	x := old[0]
	*h = old[1:]

	return x
}

const (
	heapLen = 1024
)

type broker struct {
	sync.Mutex
	eventBytes       []byte            // 暂存的 数据 ，拼接得到的结果存储再此
	eventLengthQueue *eventLengthQueue // 使用队列 记录 eventBytes 加入的每个 event 数据 的长度
	maxLen           int               // eventBytes 支持的最大字节数可能会，浮动范围取决加入的新字段的长度

	// EventSource
	subscribers map[chan []byte]bool // SSE (Server-sent Events )使用的接收者
}

// Publish sends a slice of bytes to all subscribed clients
func (b *broker) Publish(buf []byte) {
	b.Lock()
	defer b.Unlock()

	log.Debug().RawJSON("push", buf).Msg("Publish")

	for ch := range b.subscribers {
		ch <- buf
	}

	b.updateBytes()

	if len(b.eventBytes) != 0 {
		b.eventBytes = append(b.eventBytes, ',')
	}
	b.eventBytes = append(b.eventBytes, buf...)
	b.eventLengthQueue.push(len(buf))
}

func (b *broker) updateBytes() {
	for len(b.eventBytes) > b.maxLen {
		// i := getFirstPoint(b.eventBytes)// 已弃用
		// b.eventBytes = b.eventBytes[i+1:]
		i := b.eventLengthQueue.pop()
		if i >= len(b.eventBytes) {
			b.eventBytes = b.eventBytes[:0]
		} else {
			b.eventBytes = b.eventBytes[i+1:]
		}
	}
}

func (b *broker) GetAllEvents() (events []byte) {
	b.Lock()
	defer b.Unlock()
	events = b.eventBytes
	// events = make([]byte, len(h.eventBytes))
	// copy(events, h.eventBytes)
	return
}

func (b *broker) Reset() {
	b.Lock()
	defer b.Unlock()
	b.eventBytes = b.eventBytes[:0]
	b.subscribers = make(map[chan []byte]bool)
}

//////////////////////////////////// EventSource

// Subscribe adds a client to the broker
func (b *broker) Subscribe() chan []byte {
	b.Lock()
	defer b.Unlock()
	ch := make(chan []byte)
	b.subscribers[ch] = true
	return ch
}

// Unsubscribe removes a client from the broker
func (b *broker) Unsubscribe(ch chan []byte) {
	b.Lock()
	defer b.Unlock()
	close(ch)
	delete(b.subscribers, ch)
}

///////////////////////////////////

var events *broker

func init() {
	events = &broker{
		maxLen:           heapLen,
		eventBytes:       make([]byte, 0, heapLen),
		eventLengthQueue: new(eventLengthQueue),
		subscribers:      make(map[chan []byte]bool),
	}
	// heap.Init(events.eventLengthQueue)
}

func PushEvent(w http.ResponseWriter, req *http.Request) {
	resp := zerolog.New(w)
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.Err(err).Msg("req read body")
		return
	}
	if len(b) < 2 {
		resp.Err(errors.New("req body length is short")).Msg("req read body")
		return
	}
	if b[0] != '{' {
		resp.Err(errors.New("req body not json")).Msg("req body prefix")
		return
	}

	events.Publish(b)

	resp.Info().Msg("ok")
}

func GetLatestEvents(w http.ResponseWriter, req *http.Request) {
	resp := zerolog.New(w)

	data, err := getPoolEvents(nil)
	if err != nil {
		resp.Err(err).Msg("req get body")
		return
	}
	resp.Info().RawJSON("data", data).Msg("ok")
}

func getPoolEvents(_ []string) (data []byte, err error) {
	data = append(data, '[')
	b := events.GetAllEvents()
	if len(b) > 0 {
		data = append(data, b...)
	}
	data = append(data, ']')

	return data, nil
}

// https://developer.mozilla.org/zh-CN/docs/Web/API/EventSource
// 结构信息 https://javascript.ruanyifeng.com/htmlapi/eventsource.html

func EventSource(w http.ResponseWriter, req *http.Request) {
	// resp := zerolog.New(w)

	ctx := req.Context()
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")

	ch := events.Subscribe()
	defer events.Unsubscribe(ch)

	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			f.Flush()
		case <-ctx.Done():
			return
		}
	}

}

//////////////////////////////

// 通过读取 json 的结尾 } 找出最近的一个 event，存在特殊字符，
// 已弃用 使用eventLengthHeap 进行记录每个的长度
func getFirstPoint(b []byte) int {
	left := 0
	// -2 是因为 外部结果 h.eventBytes[i+1:], 防止 越界,输出的结果 往前移一位
	n := len(b) - 2
	for i, v := range b {
		/*
			// todo 处理 以下规则, 转译符号 。。。。
			{
				"a":'{}}}}}}',
				"b":"{}}}}}}"，
				"c":'{}}"}}}}'
			}
		*/
		if v == '{' {
			left++
			continue
		}
		if v == '}' {
			left--
			if left == 0 {
				if i-1 == n {
					return n
				}
				return i
			}
		}
	}
	return n
}
