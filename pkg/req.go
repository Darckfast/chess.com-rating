package chess

import (
	"bytes"
	"context"
	"sync"

	"github.com/syumai/workers/cloudflare/fetch"
)

type SendLogsArgs struct {
	Ctx      context.Context
	MaxQueue chan int
	Wg       *sync.WaitGroup
	Method   string
	Url      string
	Bearer   string
	Body     *[]byte
}

type SendLogsFunc func(args SendLogsArgs)

var SendLogs = func(args SendLogsArgs) {
	args.MaxQueue <- 1
	args.Wg.Add(1)

	req, _ := fetch.NewRequest(context.Background(), args.Method, args.Url, bytes.NewBuffer(*args.Body))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", args.Bearer)

	client := fetch.NewClient()

	go func() {
		defer args.Wg.Done()

		client.Do(req, nil)

		<-args.MaxQueue
	}()
}
