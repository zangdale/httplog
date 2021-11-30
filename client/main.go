package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	b, err := json.Marshal(struct{ Sender, Text string }{
		Sender: "Go client",
		Text:   fmt.Sprintf("The time is %s", time.Now().Format(time.Kitchen)),
	})
	check(err)
	sbu := strings.Builder{}
	sbu.WriteString(`{"type":"json","sender":"`)
	sbu.WriteString("client")
	sbu.WriteString(`","data":`)
	sbu.Write(b)
	sbu.WriteString(`}`)

	buf := bytes.NewBuffer([]byte(sbu.String()))
	resp, err := http.Post("http://localhost:80/events", "application/json", buf)
	check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	fmt.Println(string(body))
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
