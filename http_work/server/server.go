package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type appServer struct {
}

type debugServer struct {
}

func (a appServer) index(resp http.ResponseWriter, request *http.Request) {
	time.Sleep(5 * time.Second)
	resp.WriteHeader(200)
	_, err := resp.Write([]byte("Main process.\n"))
	if err != nil {
		print(err.Error())
	}
}

func (a appServer) Handler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/index", a.index)
	return mux
}

func (d debugServer) debug(resp http.ResponseWriter, request *http.Request) {
	resp.WriteHeader(200)
	_, err := resp.Write([]byte("Debug process.\n"))
	if err != nil {
		print(err.Error())
	}
}

func (d debugServer) Handler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug", d.debug)
	return mux
}

func listenSignal(errCtx context.Context, httpSrv *http.Server) error {
	quitSignal := make(chan os.Signal, 0)
	signal.Notify(quitSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-errCtx.Done():
		return errCtx.Err()
	case <-quitSignal:
		return errors.Errorf("[%s] got quit signal\n", httpSrv.Addr)
	}
}

func main() {
	appSrv := appServer{}
	debugSrv := debugServer{}
	app := &http.Server{Addr: "127.0.0.1:8080", Handler: appSrv.Handler()}
	debug := &http.Server{Addr: "127.0.0.1:8081", Handler: debugSrv.Handler()}
	serverList := []*http.Server{app, debug}

	g, errCtx := errgroup.WithContext(context.Background())
	for _, srv := range serverList {
		srvTemp := srv
		g.Go(func() error {
			fmt.Printf("[%s] http server starting...\n", srvTemp.Addr)
			return srvTemp.ListenAndServe()
		})

		g.Go(func() error {
			select {
			case <-errCtx.Done():
				fmt.Printf("[%s] http server is going to exit...\n", srvTemp.Addr)
			}
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			fmt.Printf("[%s] http server shutting down...\n", srvTemp.Addr)
			return srvTemp.Shutdown(timeoutCtx)
		})

		g.Go(func() error {
			return listenSignal(errCtx, srvTemp)
		})
	}

	err := g.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
