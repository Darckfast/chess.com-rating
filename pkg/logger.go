//go:build js && wasm

package chess

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Darckfast/workers-go/cloudflare/fetch"
)

const (
	slogFields    string = "slog_fields"
	MESSAGE_KEY   string = "msg"
	LEVEL_KEY     string = "level"
	TIMESTAMP_KEY string = "timestamp"
)

var (
	letters     = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	axiomApiKey string
	serviceName string
	maxQueue    chan int
	transport   http.RoundTripper
	wg          *sync.WaitGroup
)

type NewHandlerArgs struct {
	out         io.Writer
	serviceName string
	axiomApiKey string
}

type Handler struct {
	startedAt time.Time
	slog.Handler
	l *log.Logger
}

func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	fields := make(map[string]any, record.NumAttrs())

	fields[MESSAGE_KEY] = record.Message
	fields[LEVEL_KEY] = record.Level.String()
	fields[TIMESTAMP_KEY] = record.Time.UTC()

	record.Attrs(func(attr slog.Attr) bool {
		fields[attr.Key] = attr.Value.Any()
		return true
	})

	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		for _, attr := range attrs {
			fields[attr.Key] = attr.Value.Any()
		}
	}

	duration := time.Since(h.startedAt).Nanoseconds()
	fields["duration"] = duration

	jsonBytes, _ := json.Marshal(fields)
	h.l.Println(string(jsonBytes))
	body, _ := json.Marshal([]any{fields})

	if axiomApiKey != "" {
		sendLogOverHTTP(ctx, &body)
	}

	return nil
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func sendLogOverHTTP(ctx context.Context, body *[]byte) {
	maxQueue <- 1
	wg.Add(1)

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.axiom.co/v1/datasets/main/ingest", bytes.NewBuffer(*body))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+axiomApiKey)

	client := fetch.Client{
		Timeout: 1 * time.Second,
	}

	go func() {
		defer wg.Done()
		rs, err := client.Do(req)

		if err != nil {
			log.Println("error sending logs over http", err.Error())
		} else if rs.StatusCode > 399 {
			log.Println("axiom returned non 200 status", rs.StatusCode)
		}
		<-maxQueue
	}()
}

func NewLogger(args *NewHandlerArgs) *slog.Logger {
	return slog.New(NewHandler(args))
}

func NewHandler(
	args *NewHandlerArgs,
) *Handler {
	h := &Handler{
		startedAt: time.Now(),
		Handler:   slog.NewJSONHandler(args.out, nil),
		l:         log.New(args.out, "", 0),
	}

	axiomApiKey = args.axiomApiKey
	serviceName = args.serviceName

	return h
}

func AppendCtx(parent context.Context, attr slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(slogFields).([]slog.Attr); ok {
		v = append(v, attr)
		return context.WithValue(parent, slogFields, v)
	}

	v := []slog.Attr{}
	v = append(v, attr)
	return context.WithValue(parent, slogFields, v)
}

func FromContext(r *http.Request) (*sync.WaitGroup, *http.Request) {
	id := randSeq(24)
	ctx := AppendCtx(r.Context(), slog.String("id", id))

	if r != nil {
		ctx = AppendCtx(ctx, slog.String("query", r.URL.RawQuery))
		ctx = AppendCtx(ctx, slog.String("user-agent", r.UserAgent()))
		ctx = AppendCtx(ctx, slog.String("ip", r.RemoteAddr))
		ctx = AppendCtx(ctx, slog.String("host", r.Host))
		ctx = AppendCtx(ctx, slog.String("method", r.Method))

		requestIp := r.Header.Get("X-Forwarded-For")
		connectingIp := r.Header.Get("CF-Connecting-IP")
		if requestIp != "" && connectingIp != "" {
			requestIp += ","
		}
		requestIp += connectingIp

		ctx = AppendCtx(ctx, slog.String("x-forwarded-for", requestIp))
		ctx = AppendCtx(ctx, slog.String("country", r.Header.Get("CF-IPCountry")))
		ctx = AppendCtx(ctx, slog.String("content-type", r.Header.Get("content-type")))
		ctx = AppendCtx(ctx, slog.String("path", r.URL.Path))
	}

	ctx = AppendCtx(ctx, slog.String("service", serviceName))

	wg = &sync.WaitGroup{}
	maxQueue = make(chan int, 5)

	r = r.WithContext(ctx)

	return wg, r
}
