/*
modification history
--------------------
2018/12/21, by lovejoy, create
*/

package main

import (
	"fmt"
	"net/http"

	"context"
	"crypto/tls"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

type config struct {
	port       int
	certFile   string
	keyFile    string
	policyFile string
}

var c config

func init() {
	flag.IntVar(&c.port, "port", 443, "port")
	flag.StringVar(&c.certFile, "tls-cert-file", "server.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&c.keyFile, "tls-private-key-file", "server.key", "File containing the default x509 private key matching --tls-cert-file")
	flag.StringVar(&c.policyFile, "policy-file", "policy.json", "File containing the policy file in one json object per line format")
	flag.Parse()
	if len(policys) == 0 {
		var err error
		policys, err = newPolicyListFromFile(c.policyFile)
		if err != nil {
			panic(fmt.Sprintf("read policy file %s error : %v", c.policyFile, err))
		}
	}
}

func main() {
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
