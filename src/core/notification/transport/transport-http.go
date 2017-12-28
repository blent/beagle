package transport

import (
	"github.com/sethgrid/pester"
	"go.uber.org/zap"
	"net/http"
)

const maxConcurrency = 250

type HttpTransport struct {
	engine *pester.Client
}

func NewHttpTransport(logger *zap.Logger) *HttpTransport {
	engine := pester.New()
	engine.Backoff = pester.ExponentialBackoff
	engine.MaxRetries = 5
	engine.Concurrency = maxConcurrency
	engine.KeepLog = true
	engine.LogHook = func(e pester.ErrEntry) {
		logger.Error(
			"failed to do a request",
			zap.Error(e.Err),
			zap.String("url", e.URL),
			zap.String("method", e.Method),
			zap.String("verb", e.Verb),
			zap.Int("retry", e.Retry),
			zap.Int("attempt", e.Attempt),
		)
	}

	return &HttpTransport{
		engine,
	}
}

func (t *HttpTransport) Do(req *http.Request) error {
	_, err := t.engine.Do(req)

	return err
}
