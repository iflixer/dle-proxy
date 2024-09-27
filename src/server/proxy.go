package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func (s *Service) Proxy(w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	// get domain settings
	hostFull := strings.Split(r.Host, ":")
	host := hostFull[0]

	dom, err := s.domainService.GetDomain(host)
	//log.Printf("%+v", dom)
	if err != nil {
		log.Println(err)
		http.Error(w, "Proxy error - domain not found", http.StatusInternalServerError)
		return
	}

	targetHost := dom.ServiceDle
	if strings.HasPrefix(r.URL.String(), "/posts/") {
		targetHost = dom.ServiceImager
	}

	// Create a new HTTP request with the same method, URL, and body as the original request
	targetURL := targetHost + r.URL.String()

	log.Println(targetURL)
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	// replace host only for dle
	if targetHost == dom.ServiceDle {
		proxyReq.Host = dom.HostPrivate
	}

	// Copy the headers from the original request to the proxy request
	//log.Println("REQUEST")
	for name, values := range r.Header {
		for _, value := range values {
			//log.Println(name, value)
			if name == "Accept-Encoding" {
				value = "" // avoid gzip by backend
			}
			if isHopHeader(name) {
				continue
			}

			proxyReq.Header.Add(name, value)
		}
	}

	proxyReq.Header.Add("X-Domain-Id", fmt.Sprintf("%d", dom.ID))

	//Send the proxy request using the custom transport
	resp, err := s.customTransport.RoundTrip(proxyReq)
	if err != nil {
		log.Println(err)
		http.Error(w, "Proxy error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy the headers from the proxy response to the original response
	needReplace := false
	for name, values := range resp.Header {
		for _, value := range values {
			if name == "X-Powered-By" {
				continue
			}
			if name == "Server" {
				continue
			}
			// need to modify html
			if name == "Content-Type" && strings.HasPrefix(value, "text/html") {
				needReplace = true
			}
			//log.Println("response header:", name, value)
			w.Header().Add(name, value)
		}
	}

	if needReplace {
		body, _ := io.ReadAll(resp.Body)

		pubURL := dom.SchemePublic + "://" + dom.HostPublic
		if dom.PortPublic != "" {
			pubURL += ":" + dom.PortPublic
		}

		body = bytes.ReplaceAll(body, []byte("https://"+dom.HostPrivate), []byte(pubURL))
		body = bytes.ReplaceAll(body, []byte("http://"+dom.HostPrivate), []byte(pubURL))

		// remove S3 domain for images
		body = bytes.ReplaceAll(body, []byte(dom.ServiceImager), []byte(""))

		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))

		w.Header().Add("X-Proxy-tm", fmt.Sprintf("%d", time.Since(start).Milliseconds()))
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	w.Header().Add("X-Proxy-tm", fmt.Sprintf("%d", time.Since(start).Milliseconds()))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}

func isHopHeader(header string) bool {
	for _, h := range hopHeaders {
		if header == h {
			return true
		}
	}
	return false
}
