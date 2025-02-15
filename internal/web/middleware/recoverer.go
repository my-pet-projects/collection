package middleware

// The original work was derived from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/web"
)

// Recoverer is a middleware that recovers from panics, logs the panic (and a
// backtrace), and returns a HTTP 500 (Internal Server Error) status if
// possible. Recoverer prints a request ID if one is provided.
//
// Alternatively, look at https://github.com/go-chi/httplog middleware pkgs.

func (m Middleware) WithRecoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqResp := &web.ReqRespPair{
			Response: w,
			Request:  r,
		}

		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				m.logger.Error("Panic", slog.Any("panic", rvr), slog.Any("stack", string(debug.Stack())))

				printPrettyStack(rvr)

				var panicErr error
				switch typedErr := rvr.(type) {
				case string:
					panicErr = errors.New(typedErr)
				case error:
					panicErr = errors.Wrap(typedErr, "panic")
				default:
					panicErr = errors.New("panic has occurred")
				}

				reqResp.RenderErrorPage(http.StatusInternalServerError, panicErr)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// for ability to test the PrintPrettyStack function.
var recovererErrorWriter io.Writer = os.Stderr

func printPrettyStack(rvr interface{}) {
	debugStack := debug.Stack()
	s := prettyStack{}
	out, err := s.parse(debugStack, rvr)
	if err == nil {
		recovererErrorWriter.Write(out)
	} else {
		// print stdlib output as a fallback
		os.Stderr.Write(debugStack)
	}
}

type prettyStack struct{}

func (s prettyStack) parse(debugStack []byte, rvr interface{}) ([]byte, error) {
	var err error
	useColor := true
	buf := &bytes.Buffer{}

	colorWrite(buf, false, bRed, "\n")
	colorWrite(buf, useColor, bCyan, " panic: ")
	colorWrite(buf, useColor, bBlue, "%v", rvr)
	colorWrite(buf, false, bWhite, "\n \n")

	// process debug stack info
	stack := strings.Split(string(debugStack), "\n")
	lines := []string{}

	// locate panic line, as we may have nested panics
	for i := len(stack) - 1; i > 0; i-- {
		lines = append(lines, stack[i])
		if strings.HasPrefix(stack[i], "panic(") {
			lines = lines[0 : len(lines)-2] // remove boilerplate
			break
		}
	}

	// reverse
	for i := len(lines)/2 - 1; i >= 0; i-- { //nolint:mnd
		opp := len(lines) - 1 - i
		lines[i], lines[opp] = lines[opp], lines[i]
	}

	// decorate
	for i, line := range lines {
		lines[i], err = s.decorateLine(line, useColor, i)
		if err != nil {
			return nil, err
		}
	}

	for _, l := range lines {
		fmt.Fprintf(buf, "%s", l)
	}
	return buf.Bytes(), nil
}

func (s prettyStack) decorateLine(line string, useColor bool, num int) (string, error) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "\t") || strings.Contains(line, ".go:") {
		return s.decorateSourceLine(line, useColor, num)
	}
	if strings.HasSuffix(line, ")") {
		return s.decorateFuncCallLine(line, useColor, num)
	}
	if strings.HasPrefix(line, "\t") {
		return strings.Replace(line, "\t", "      ", 1), nil
	}
	return fmt.Sprintf("    %s\n", line), nil
}

func (s prettyStack) decorateFuncCallLine(line string, useColor bool, num int) (string, error) {
	idx := strings.LastIndex(line, "(")
	if idx < 0 {
		return "", errors.New("not a func call line")
	}

	buf := &bytes.Buffer{}
	pkg := line[0:idx]
	// addr := line[idx:]
	method := ""

	if idx := strings.LastIndex(pkg, string(os.PathSeparator)); idx < 0 {
		if idx := strings.Index(pkg, "."); idx > 0 {
			method = pkg[idx:]
			pkg = pkg[0:idx]
		}
	} else {
		method = pkg[idx+1:]
		pkg = pkg[0 : idx+1]
		if idx := strings.Index(method, "."); idx > 0 {
			pkg += method[0:idx]
			method = method[idx:]
		}
	}
	pkgColor := nYellow
	methodColor := bGreen

	if num == 0 {
		colorWrite(buf, useColor, bRed, " -> ")
		pkgColor = bMagenta
		methodColor = bRed
	} else {
		colorWrite(buf, useColor, bWhite, "    ")
	}
	colorWrite(buf, useColor, pkgColor, "%s", pkg)
	colorWrite(buf, useColor, methodColor, "%s\n", method)
	// cW(buf, useColor, nBlack, "%s", addr)
	return buf.String(), nil
}

func (s prettyStack) decorateSourceLine(line string, useColor bool, num int) (string, error) {
	idx := strings.LastIndex(line, ".go:")
	if idx < 0 {
		return "", errors.New("not a source line")
	}

	buf := &bytes.Buffer{}
	path := line[0 : idx+3]
	lineno := line[idx+3:]

	idx = strings.LastIndex(path, string(os.PathSeparator))
	dir := path[0 : idx+1]
	file := path[idx+1:]

	idx = strings.Index(lineno, " ")
	if idx > 0 {
		lineno = lineno[0:idx]
	}
	fileColor := bCyan
	lineColor := bGreen

	if num == 1 {
		colorWrite(buf, useColor, bRed, " ->   ")
		fileColor = bRed
		lineColor = bMagenta
	} else {
		colorWrite(buf, false, bWhite, "      ")
	}
	colorWrite(buf, useColor, bWhite, "%s", dir)
	colorWrite(buf, useColor, fileColor, "%s", file)
	colorWrite(buf, useColor, lineColor, "%s", lineno)
	if num == 1 {
		colorWrite(buf, false, bWhite, "\n")
	}
	colorWrite(buf, false, bWhite, "\n")

	return buf.String(), nil
}

func colorWrite(w io.Writer, useColor bool, color []byte, s string, args ...interface{}) {
	if useColor {
		w.Write(color)
	}
	fmt.Fprintf(w, s, args...)
	if useColor {
		w.Write(reset)
	}
}

var (
	// Normal colors.
	nBlack   = []byte{'\033', '[', '3', '0', 'm'}
	nRed     = []byte{'\033', '[', '3', '1', 'm'}
	nGreen   = []byte{'\033', '[', '3', '2', 'm'}
	nYellow  = []byte{'\033', '[', '3', '3', 'm'}
	nBlue    = []byte{'\033', '[', '3', '4', 'm'}
	nMagenta = []byte{'\033', '[', '3', '5', 'm'}
	nCyan    = []byte{'\033', '[', '3', '6', 'm'}
	nWhite   = []byte{'\033', '[', '3', '7', 'm'}
	// Bright colors.
	bBlack   = []byte{'\033', '[', '3', '0', ';', '1', 'm'}
	bRed     = []byte{'\033', '[', '3', '1', ';', '1', 'm'}
	bGreen   = []byte{'\033', '[', '3', '2', ';', '1', 'm'}
	bYellow  = []byte{'\033', '[', '3', '3', ';', '1', 'm'}
	bBlue    = []byte{'\033', '[', '3', '4', ';', '1', 'm'}
	bMagenta = []byte{'\033', '[', '3', '5', ';', '1', 'm'}
	bCyan    = []byte{'\033', '[', '3', '6', ';', '1', 'm'}
	bWhite   = []byte{'\033', '[', '3', '7', ';', '1', 'm'}

	reset = []byte{'\033', '[', '0', 'm'}
)
