package chess

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/syumai/workers/cloudflare"
)

const (
	slogFields     string = "slog_fields"
	REQUEST_ID_KEY string = "requestId"
	NAMESPACE_KEY  string = "namespace"
	MESSAGE_KEY    string = "message"
	LEVEL_KEY      string = "level"
	TIMESTAMP_KEY  string = "timestamp"
	STARTED_AT_KEY string = "startedAt"
	DATA_KEY       string = "data"
)

var (
	AXIOM_API_KEY = ""
	SERVICE_NAME  = ""
)

type Handler struct {
	slog.Handler
	l *log.Logger
}

var sendLogsArs SendLogsArgs

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

	if fields[STARTED_AT_KEY] == nil {
		timeNow := time.Now()
		fields[STARTED_AT_KEY] = timeNow
		AppendCtx(ctx, slog.Time(STARTED_AT_KEY, timeNow))
	}

	if fields["duration"] == nil {
		startedAt := fields[STARTED_AT_KEY].(time.Time)
		duration := time.Since(startedAt).Milliseconds()
		fields["duration"] = duration
	}

	jsonBytes, _ := json.Marshal(fields)
	h.l.Println(string(jsonBytes))
	body, _ := json.Marshal([]any{fields})

	if AXIOM_API_KEY != "" {
		sendLogsArs.Body = &body
		SendLogs(sendLogsArs)
	}

	return nil
}

func NewHandler(
	out io.Writer,
) *Handler {
	h := &Handler{
		Handler: slog.NewJSONHandler(out, nil),
		l:       log.New(out, "", 0),
	}

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

type SetupOps struct {
	AxiomApiKey string
	ServiceName string
	Request     *http.Request
	RequestGen  SendLogsFunc
}

func SetupContext(r *http.Request) *sync.WaitGroup {
	uid, _ := uuid.NewV7()

	ctx := AppendCtx(r.Context(), slog.String(REQUEST_ID_KEY, uid.String()))

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
		ctx = AppendCtx(ctx, slog.Int64("content-length", r.ContentLength))
		ctx = AppendCtx(ctx, slog.String("content-type", r.Header.Get("content-type")))
		ctx = AppendCtx(ctx, slog.String(NAMESPACE_KEY, r.URL.Path))
	}

	ctx = AppendCtx(ctx, slog.String("service", "chess-com-rating"))
	ctx = AppendCtx(ctx, slog.Time(STARTED_AT_KEY, time.Now()))

	AXIOM_API_KEY = cloudflare.Getenv("AXIOM_API_KEY")

	sendLogsArs.Wg = &sync.WaitGroup{}
	sendLogsArs.Ctx = ctx
	sendLogsArs.MaxQueue = make(chan int, 5)
	sendLogsArs.Method = "POST"
	sendLogsArs.Url = "https://api.axiom.co/v1/datasets/main/ingest"
	sendLogsArs.Bearer = "Bearer " + AXIOM_API_KEY

	r = r.WithContext(ctx)

	return sendLogsArs.Wg
}
