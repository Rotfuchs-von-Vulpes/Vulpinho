package log

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
)

type CopyHandler struct {
	mu  *sync.Mutex
	out []slog.Handler // all the destinations
}

func NewCopyHandler(handlers ...slog.Handler) *CopyHandler {
	h := &CopyHandler{out: handlers, mu: &sync.Mutex{}}
	return h
}

func (h *CopyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// leave the enable check to the underlying handlers
	return true
}

func (h *CopyHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, destHandler := range h.out {
		err := destHandler.Handle(ctx, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *CopyHandler) WithGroup(name string) slog.Handler {
	// call WithGroup on the underlying handlers
	// we should not make modification the receiver, we return a copy
	if name == "" {
		return h
	}
	h2 := *h
	h2.out = make([]slog.Handler, len(h.out))
	for i, h := range h.out {
		h2.out[i] = h.WithGroup(name)
	}
	return &h2
}

func (h *CopyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// call WithAttrs on the underlying handlers
	// we should not make modification the receiver, we return a copy
	if len(attrs) == 0 {
		return h
	}
	h2 := *h
	h2.out = make([]slog.Handler, len(h.out))
	for i, h := range h.out {
		h2.out[i] = h.WithAttrs(attrs)
	}
	return &h2
}

func fileExist(path string) (bool, bool) {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, false
	} else if err == nil {
		return true, false
	} else {
		slog.Error("error viewing stat", "err", err)
		return false, true
	}
}

func removeAndCreate(path string) (*os.File, bool) {
	exist, fatal := fileExist(path)
	if fatal {
		return nil, true
	}
	if exist {
		if err := os.Remove(path); err != nil {
			slog.Error("error removing file", "err", err.Error())
			return nil, true
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE, os.ModeDevice)
	if err != nil {
		slog.Error("error opening file", "err", err.Error())
		return nil, true
	}
	return f, false
}

var logFile *os.File

func Init() bool {
	fatal := false
	logFile, fatal = removeAndCreate("log/log.txt")
	if fatal {
		return true
	}
	logger := slog.New(slog.NewMultiHandler(slog.NewTextHandler(os.Stdout, nil), slog.NewJSONHandler(logFile, nil)))
	slog.SetDefault(logger)
	return false
}

func End() {
	logFile.Close()
}
