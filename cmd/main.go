package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
)

var (
	Version   string
	GoVersion string
	GitCommit string
	BuildTime string
)

func versionInfo() {
	fmt.Printf("Version:\t%s\n", Version)
	fmt.Printf("Go version:\t%s\n", GoVersion)
	fmt.Printf("Git commit:\t%s\n", GitCommit)
	fmt.Printf("Built:\t\t%s\n", BuildTime)
}

func startWebhookServer(parameters *WhSvrParameters) *WebhookServer {
	pair, err := tls.LoadX509KeyPair(parameters.certFile, parameters.keyFile)
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}

	whsvr := &WebhookServer{
		server: &http.Server{
			Addr:      fmt.Sprintf(":%v", parameters.port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", whsvr.ping)
	mux.HandleFunc("/mutate", whsvr.serve)
	whsvr.server.Handler = mux

	// start webhook server in new rountine
	go func() {
		if err := whsvr.server.ListenAndServeTLS("", ""); err != nil {
			glog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	return whsvr
}

func main() {
	var parameters WhSvrParameters
	var echoVersion bool

	// get command line parameters
	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
	flag.BoolVar(&echoVersion, "version", false, "Show the LXCFS admission webhook version information")
	flag.Parse()

	defer glog.Flush()

	if echoVersion {
		versionInfo()
		os.Exit(0)
	}

	whsvr := startWebhookServer(&parameters)

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	if err := whsvr.server.Shutdown(context.Background()); err != nil {
		glog.Errorf("Errors when shutting service: %v", err)
	}
}
