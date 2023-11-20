package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", 4000),
		Handler: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			},
		),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownErr := make(chan error)
	go func() {
		defer close(shutdownErr)
		quit := make(chan os.Signal, 1)
		signal.Notify(quit,
			os.Interrupt,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)
		signalReceived := <-quit
		fmt.Printf("Received signal: %v, shutting down gracefully...", signalReceived)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			shutdownErr <- err
		} else {
			fmt.Print("Server gracefully shut down")
		}
	}()

	fmt.Printf("Server is now listening")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := <-shutdownErr; err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
