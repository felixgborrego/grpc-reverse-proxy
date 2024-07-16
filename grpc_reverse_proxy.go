package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type SidecarReverseProxy struct {
	TargetUrl *url.URL
	Authority string
	logger    *zap.SugaredLogger
}

const (
	proxyPort = ":8083"
)

func NewSidecarReverseProxy(logger *zap.SugaredLogger, proxyTarget string, proxyAuthority string) (*SidecarReverseProxy, error) {
	targetUrl, err := url.Parse(proxyTarget)
	if err != nil {
		log.Fatalf("invalid target URL: %s", proxyTarget)
	}

	if proxyAuthority == "" {
		proxyAuthority = targetUrl.Host
		if targetUrl.Port() != "" {
			proxyAuthority += ":" + targetUrl.Port()
		}
		logger.Infof("Grpc Authority not set, defaulting to target URL authority: %s", proxyAuthority)
	}

	return &SidecarReverseProxy{
		TargetUrl: targetUrl,
		Authority: proxyAuthority,
		logger:    logger,
	}, nil
}

func (s *SidecarReverseProxy) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	s.logger.Infof("Request received: %s %s to %s (Authority: %s)", req.Method, req.URL.Path, s.TargetUrl.String(), s.Authority)

	// Create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(s.TargetUrl)

	// Update the director to modify request headers to target
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		r.Header.Set("Host", s.Authority) // Ensure the Host header is set
		req.Header.Set(":authority", s.Authority)
		s.logger.Debugf("Forwarding request: %s %s -> %s", r.Method, r.URL, s.TargetUrl.String())
	}

	// Create an HTTP/2 Transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Skip verification for demonstration purposes, not recommended for production
		},
	}

	// Configure HTTP/2 for the transport
	if err := http2.ConfigureTransport(transport); err != nil {
		log.Fatalf("Failed to configure HTTP/2 transport: %v", err)
	}

	// Create a custom HTTP client using the HTTP/2 transport
	proxy.Transport = transport

	proxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, e error) {
		s.logger.Warnf("Error handling request: %v", e)
		http.Error(rw, "Error contacting backend service", http.StatusBadGateway)
	}

	// Handle the request with the reverse proxy
	req2 := req.Clone(req.Context())
	req2.TLS = nil
	// Explicitly set the scheme to "http"
	req2.URL.Scheme = "https"
	// Set the TLS field to nil
	req2.TLS = nil
	proxy.ServeHTTP(res, req2)
}

func StartReverseProxy(logger *zap.SugaredLogger, proxyTarget string, proxyAuthority string, certFile string, keyFile string) {
	sidecarProxy, err := NewSidecarReverseProxy(logger, proxyTarget, proxyAuthority)
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
		panic(err)
	}
	// Initialize HTTP Server
	server := &http.Server{
		Addr: proxyPort,
		Handler: http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			sidecarProxy.handleRequestAndRedirect(res, req)
		}),
	}

	// Enable HTTP/2
	http2.ConfigureServer(server, &http2.Server{})
	logger.Infof("Starting proxy server at %s to %s with authority %s (tls file at %s)", proxyPort, sidecarProxy.TargetUrl.String(), sidecarProxy.Authority, certFile)
	err = server.ListenAndServeTLS(certFile, keyFile)

	if err != nil {
		logger.Fatalf("Error starting server: %s", err)
		panic(err)
	}
}
