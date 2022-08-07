package log_test

import (
	"fmt"
	"testing"

	"github.com/bep/log"
	"github.com/bep/log/handlers/discard"
	"github.com/bep/log/handlers/memory"
	qt "github.com/frankban/quicktest"
)

func TestLogger_printf(t *testing.T) {
	h := memory.New()

	l := &log.Logger{
		Handler: h,
		Level:   log.InfoLevel,
	}

	l.Infof("logged in %s", "Tobi")

	e := h.Entries[0]
	qt.Assert(t, "logged in Tobi", qt.Equals, e.Message)
	qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
}

func TestLogger_levels(t *testing.T) {
	h := memory.New()

	l := &log.Logger{
		Handler: h,
		Level:   log.InfoLevel,
	}

	l.Debug("uploading")
	l.Info("upload complete")

	qt.Assert(t, len(h.Entries), qt.Equals, 1)

	e := h.Entries[0]
	qt.Assert(t, "upload complete", qt.Equals, e.Message)
	qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
}

func TestLogger_WithFields(t *testing.T) {
	h := memory.New()

	l := &log.Logger{
		Handler: h,
		Level:   log.InfoLevel,
	}

	ctx := l.WithFields(log.Fields{{"file", "sloth.png"}})
	ctx.Debug("uploading")
	ctx.Info("upload complete")

	qt.Assert(t, len(h.Entries), qt.Equals, 1)

	e := h.Entries[0]
	qt.Assert(t, "upload complete", qt.Equals, e.Message)
	qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
	qt.Assert(t, e.Fields, qt.DeepEquals, log.Fields{{"file", "sloth.png"}})
}

func TestLogger_WithField(t *testing.T) {
	h := memory.New()

	l := &log.Logger{
		Handler: h,
		Level:   log.InfoLevel,
	}

	ctx := l.WithField("file", "sloth.png").WithField("user", "Tobi")
	ctx.Debug("uploading")
	ctx.Info("upload complete")

	qt.Assert(t, len(h.Entries), qt.Equals, 1)

	e := h.Entries[0]
	qt.Assert(t, "upload complete", qt.Equals, e.Message)
	qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
	qt.Assert(t, e.Fields, qt.DeepEquals, log.Fields{{"file", "sloth.png"}, {"user", "Tobi"}})
}

func TestLogger_Trace_info(t *testing.T) {
	h := memory.New()

	l := &log.Logger{
		Handler: h,
		Level:   log.InfoLevel,
	}

	func() (err error) {
		defer l.WithField("file", "sloth.png").Trace("upload").Stop(&err)
		return nil
	}()

	qt.Assert(t, len(h.Entries), qt.Equals, 2)

	{
		e := h.Entries[0]
		qt.Assert(t, "upload", qt.Equals, e.Message)
		qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
		qt.Assert(t, e.Fields, qt.DeepEquals, log.Fields{{"file", "sloth.png"}})
	}

	{
		e := h.Entries[1]
		qt.Assert(t, "upload", qt.Equals, e.Message)
		qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
		qt.Assert(t, e.Fields[0].Value, qt.Equals, "sloth.png")
		qt.Assert(t, e.Fields[1].Value, qt.Equals, int64(0))
	}
}

func TestLogger_Trace_error(t *testing.T) {
	h := memory.New()

	l := &log.Logger{
		Handler: h,
		Level:   log.InfoLevel,
	}

	func() (err error) {
		defer l.WithField("file", "sloth.png").Trace("upload").Stop(&err)
		return fmt.Errorf("boom")
	}()

	qt.Assert(t, len(h.Entries), qt.Equals, 2)

	{
		e := h.Entries[0]
		qt.Assert(t, "upload", qt.Equals, e.Message)
		qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
		qt.Assert(t, e.Fields, qt.DeepEquals, log.Fields{
			{
				Name:  "file",
				Value: "sloth.png",
			},
		})
	}

	{
		e := h.Entries[1]
		qt.Assert(t, "upload", qt.Equals, e.Message)
		qt.Assert(t, log.ErrorLevel, qt.Equals, e.Level)
		qt.Assert(t, e.Fields, qt.DeepEquals, log.Fields{
			{
				Name:  "file",
				Value: "sloth.png",
			},
			{
				Name:  "duration",
				Value: int64(0),
			},
			{
				Name:  "error",
				Value: "boom",
			},
		},
		)

	}
}

func TestLogger_Trace_nil(t *testing.T) {
	h := memory.New()

	l := &log.Logger{
		Handler: h,
		Level:   log.InfoLevel,
	}

	func() {
		defer l.WithField("file", "sloth.png").Trace("upload").Stop(nil)
	}()

	qt.Assert(t, len(h.Entries), qt.Equals, 2)

	{
		e := h.Entries[0]
		qt.Assert(t, "upload", qt.Equals, e.Message)
		qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
		qt.Assert(t, e.Fields, qt.DeepEquals, log.Fields{{"file", "sloth.png"}})
	}

	{
		e := h.Entries[1]
		qt.Assert(t, "upload", qt.Equals, e.Message)
		qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
		qt.Assert(t, e.Fields, qt.DeepEquals, log.Fields{{Name: "file", Value: string("sloth.png")}, {Name: "duration", Value: int64(0)}})

	}
}

func TestLogger_HandlerFunc(t *testing.T) {
	h := memory.New()
	f := func(e *log.Entry) error {
		return h.HandleLog(e)
	}

	l := &log.Logger{
		Handler: log.HandlerFunc(f),
		Level:   log.InfoLevel,
	}

	l.Infof("logged in %s", "Tobi")

	e := h.Entries[0]
	qt.Assert(t, "logged in Tobi", qt.Equals, e.Message)
	qt.Assert(t, log.InfoLevel, qt.Equals, e.Level)
}

func BenchmarkLogger_small(b *testing.B) {
	l := &log.Logger{
		Handler: discard.New(),
		Level:   log.InfoLevel,
	}

	for i := 0; i < b.N; i++ {
		l.Info("login")
	}
}

func BenchmarkLogger_medium(b *testing.B) {
	l := &log.Logger{
		Handler: discard.New(),
		Level:   log.InfoLevel,
	}

	for i := 0; i < b.N; i++ {
		l.WithFields(log.Fields{
			{"file", "sloth.png"},
			{"type", "image/png"},
			{"size", 1 << 20},
		}).Info("upload")
	}
}

func BenchmarkLogger_large(b *testing.B) {
	l := &log.Logger{
		Handler: discard.New(),
		Level:   log.InfoLevel,
	}

	err := fmt.Errorf("boom")

	for i := 0; i < b.N; i++ {
		l.WithFields(log.Fields{
			{"file", "sloth.png"},
			{"type", "image/png"},
			{"size", 1 << 20},
		}).
			WithFields(log.Fields{
				{"some", "more"},
				{"data", "here"},
				{"whatever", "blah blah"},
				{"more", "stuff"},
				{"context", "such useful"},
				{"much", "fun"},
			}).
			WithError(err).Error("upload failed")
	}
}
