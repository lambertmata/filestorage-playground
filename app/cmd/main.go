package main

import (
	"fmt"
	"fsp"
	"github.com/lambertmata/churro"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"
	"time"
)

type WrappedResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *WrappedResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func NewWrappedResponseWriter(w http.ResponseWriter) *WrappedResponseWriter {
	return &WrappedResponseWriter{ResponseWriter: w, StatusCode: 200}
}

func LogRequests() churro.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()

			responseWriter := NewWrappedResponseWriter(w)

			next.ServeHTTP(responseWriter, req)

			duration := time.Since(start)

			slog.Info(fmt.Sprintf("%s %s %s %d %dms", req.RemoteAddr, req.Method, req.URL, responseWriter.StatusCode, duration.Milliseconds()))
		})
	}
}

func main() {

	r := churro.NewRouter()

	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	storagePath := path.Join(cwd, "storage")

	if err := os.RemoveAll(storagePath); err != nil {
		panic(err)
	}

	if err := os.Mkdir(storagePath, os.ModePerm); err != nil {
		panic(err)
	}

	storage := fsp.NewLocalStorage(storagePath)
	objects := fsp.NewObjectsHandler(storage)

	r.Middlewares(LogRequests())
	churro.Post(r, "/objects", objects.Put)
	churro.Get(r, "/objects/{id}", objects.Get)
	churro.Get(r, "/objects", objects.List)

	port := "8080"

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	addr := net.JoinHostPort("127.0.0.1", port)

	slog.Info("Starting file storage server at " + addr)
	http.ListenAndServe(addr, r)
}
