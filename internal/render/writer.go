package render

import (
	"fmt"
	"io"
)

type writer struct {
	w   io.Writer
	err error
}

func (w *writer) printf(format string, args ...any) {
	if w.err != nil {
		return
	}

	_, w.err = fmt.Fprintf(w.w, format, args...)
}

func (w *writer) println(args ...any) {
	if w.err != nil {
		return
	}

	_, w.err = fmt.Fprintln(w.w, args...)
}
