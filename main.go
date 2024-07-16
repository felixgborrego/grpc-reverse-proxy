package main

import (
	"os"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()

	proxyTarget := os.Getenv("PROXY_TARGET")
	if proxyTarget == "" {
		log.Fatalf("Environment variable PROXY_TARGET is required but not set")
	}

	proxyAuthority := os.Getenv("PROXY_AUTHORITY")
	if proxyAuthority == "" {
		proxyAuthority = proxyTarget
	}

	certFile := os.Getenv("TLS_CERT_FILE")
	if certFile == "" {
		certFile = "/opt/ssl/sidecar.pem"
		log.Infof("TLS_CERT_FILE not set, defaulting to /opt/ssl/sidecar.pem")
	}

	keyFile := os.Getenv("TLS_KEY_FILE")
	if keyFile == "" {
		keyFile = "/opt/ssl/sidecar.key"
		log.Infof("TLS_KEY_FILE not set, defaulting to /opt/ssl/sidecar.key")
	}

	log.Infof("Proxy Target: %s", proxyTarget)
	log.Infof("Proxy Authority: %s", proxyAuthority)
	log.Infof("TLS Cert File: %s", certFile)
	log.Infof("TLS Key File: %s", keyFile)

	StartReverseProxy(log, proxyTarget, proxyAuthority, certFile, keyFile)
}
