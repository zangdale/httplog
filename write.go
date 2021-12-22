package httplog

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var _ io.Writer = (*Writer)(nil)

type Writer func([]byte) (int, error)

func (w Writer) Write(buf []byte) (int, error) {
	return w(buf)
}

var cfg = new(conf)
var Client = new(http.Client)

type conf struct {
	url         string
	open        bool
	ignoreError bool
}

func SetOpen(b bool) {
	cfg.open = b
}

func SetIgnoreError(b bool) {
	cfg.ignoreError = b
}

func WriteJson(sender, url string) Writer {
	cfg.url = url
	return hWriteJson(sender)
}

func WriteString(sender, url string) Writer {
	cfg.url = url
	return hWriteString(sender)
}

func hWriteJson(sender string) Writer {
	sbu := strings.Builder{}
	sbu.WriteString(`{"type":"json","sender":"`)
	sbu.WriteString(sender)
	sbu.WriteString(`","data":`)
	sender = sbu.String()
	return func(b []byte) (int, error) {
		i := len(b)
		if !cfg.open {
			return i, nil
		}
		bf := pool.Get().(*bytes.Buffer)
		bf.WriteString(sender)

		bf.Write(bytes.TrimFunc(b, func(r rune) bool {
			if r == '\t' || r == '\n' {
				return true
			}
			return false
		}))

		bf.WriteString(`}`)
		return i, push(cfg, bf)
	}
}

func hWriteString(sender string) Writer {
	// 其他 可以直接
	/*
		bf:=bytes.Buffer{}
		resp := zerolog.New(w)
		//resp.Info().RawJSON("data", data).Msg("ok")
		resp.Info().Bytes("data", data).Msg("ok")
	*/
	sbu := strings.Builder{}
	sbu.WriteString(`{"type":"string","sender":"`)
	sbu.WriteString(sender)
	sbu.WriteString(`","data":"`)
	sender = sbu.String()
	return func(b []byte) (int, error) {
		i := len(b)
		if !cfg.open {
			return i, nil
		}
		bf := pool.Get().(*bytes.Buffer)
		bf.WriteString(sender)
		bf.Write(bytes.TrimFunc(b, func(r rune) bool {
			if r == '\t' || r == '\n' {
				return true
			}
			return false
		}))
		bf.WriteString(`"}`)
		return i, push(cfg, bf)
	}
}

var pool sync.Pool

func init() {
	pool.New = func() interface{} {
		return &bytes.Buffer{}
	}
}

func push(cfg *conf, bf *bytes.Buffer) (err error) {
	if bf == nil {
		return nil
	}
	fmt.Println(bf.String())
	if cfg.ignoreError {
		_, _ = Client.Post(cfg.url, "application/json", bf)
		return
	}
	_, err = Client.Post(cfg.url, "application/json", bf)
	return
}
