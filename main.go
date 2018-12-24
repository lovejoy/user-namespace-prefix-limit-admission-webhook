/*
modification history
--------------------
2018/12/21, by lovejoy, create
*/

package main

import (
	"flag"
	"fmt"
	"net/http"

	"context"
	"crypto/tls"
	"os"
	"os/signal"
	"syscall"
)

type config struct {
	port     int
	certFile string
	keyFile  string
}

func main() {
	var c config
	flag.IntVar(&c.port, "port", 443, "port")
	flag.StringVar(&c.certFile, "tls-cert-file", "server.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&c.keyFile, "tls-private-key-file", "server.key", "File containing the default x509 private key matching --tls-cert-file")
	flag.Parse()

	pair, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		fmt.Errorf("Failed to load key pair: %v", err)
	}
	http.HandleFunc("/validate", validate)

	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", c.port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
	}
	go func() {
		server.ListenAndServeTLS("", "")
	}()
	fmt.Println("webhook started")

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	server.Shutdown(context.Background())
	fmt.Println("webhook shutdowned")

}
