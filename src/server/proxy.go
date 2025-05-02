package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	//"Connection",
	//"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	//"Upgrade",
}

func (s *Service) Proxy(w http.ResponseWriter, r *http.Request) {

	//log.Println(r.URL.String())
	// r.URL.String() - with ?qwe=1
	// r.URL.Path - without ?qwe=1
	// r.URL.Query() - map[string][]string

	start := time.Now()

	forbiddenReplaceDomain := false

	// get domain settings
	// r.Host with port like proxy2.cis-dle.orb.local:8090
	hostFull := strings.Split(r.Host, ":")
	hostHeader := r.Header.Get("X-Forwarded-Host")
	if hostHeader != "" {
		hostFull = strings.Split(hostHeader, ":")
	}
	host := hostFull[0]
	path := r.URL.Path
	uri := r.URL.String()

	//log.Println(host, path)

	// check if this domain is alias so we need to redirect to main domain
	alias, err := s.domainAliasService.GetDomain(host)
	if err == nil {
		log.Printf("domain alias %s id %d\n", host, alias.DomainID)
		domain, err := s.domainService.GetDomainByID(alias.DomainID)
		if err == nil {
			targetURL := fmt.Sprintf("https://%s%s", domain.HostPublic, uri)
			log.Printf("%s 302 %s\n", uri, targetURL)
			http.Redirect(w, r, targetURL, http.StatusMovedPermanently)
			return
		}
	}

	dom, err := s.domainService.GetDomain(host)
	if err != nil {
		log.Println("Proxy error - domain ["+host+"] not found", err)
		http.Error(w, "Proxy error - domain ["+host+"] not found", http.StatusNotFound)
		return
	}

	if strings.HasPrefix(uri, "/robots.txt") && dom.DisallowRobots {
		w.Write([]byte(`User-agent: *
Disallow: /

User-agent: YandexBot
Disallow: /

Host: https://` + host + `/`))
		return
	}

	// file request?
	if file, err := s.fileService.GetFile(dom.ID, path); err == nil {
		log.Printf("%s STAT\n", path)
		w.Header().Set("Content-Type", file.ContentType)
		w.Write([]byte(file.Body))
		return
	}

	// check if we have url overrides in flix_post
	if strings.HasSuffix(path, ".html") {
		post, altName, err := s.flixPostService.GetPost(dom.ID, path)
		if err == nil {
			// we have override
			if post.AltName != altName {
				if post.Redirect == 1 {
					targetURI := strings.Replace(uri, altName+".html", post.AltName+".html", 1)
					targetURL := fmt.Sprintf("https://%s%s", dom.HostPublic, targetURI)
					log.Printf("%s 301 %s\n", uri, targetURL)
					w.Header().Set("X-Proxy-Redirect-Reason", "fdjiehfueig37367")
					http.Redirect(w, r, targetURL, http.StatusMovedPermanently)
					return
				} else {
					log.Println("StatusPaymentRequired " + uri)
					w.Header().Set("X-Proxy-Redirect-Reason", "vedfdsfd323ddd")
					w.WriteHeader(http.StatusUnavailableForLegalReasons)
					return
				}
			}
		}
	}

	targetHost := dom.ServiceDle
	if strings.HasPrefix(uri, "/posts/") || strings.HasPrefix(uri, "/fotos/") {
		targetHost = dom.ServiceImager
		forbiddenReplaceDomain = true
	}

	if strings.HasPrefix(uri, "/stater/") {
		targetHost = "http://stater"
		forbiddenReplaceDomain = true
	}

	if strings.HasPrefix(uri, "/resize/") || strings.HasPrefix(uri, "/crop/") {
		targetHost = "http://imaginary:8088"
		//path = strings.ReplaceAll(path, "/resize/", "/crop/")
		uri = strings.ReplaceAll(uri, "?w=", "?width=")
		uri = strings.ReplaceAll(uri, "?h=", "?height=")
		uri = strings.ReplaceAll(uri, "&w=", "&width=")
		uri = strings.ReplaceAll(uri, "&h=", "&height=")
		forbiddenReplaceDomain = true
	}

	if strings.HasPrefix(uri, "/sitemap") {
		targetHost = dom.ServiceSitemap
		forbiddenReplaceDomain = true
	}

	if path == "/traefik" {
		targetHost = dom.ServiceDns
		forbiddenReplaceDomain = true
	}

	// Create a new HTTP request with the same method, URL, and body as the original request
	targetURL := targetHost + uri

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		log.Println(err)
		log.Println("Error creating proxy request", err.Error())
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
			if name == "Referer" {
				value = strings.ReplaceAll(value, host, dom.HostPrivate)
			}
			if isHopHeader(name) {
				continue
			}

			proxyReq.Header.Add(name, value)
		}
	}

	// allow cloudflare cache

	proxyReq.Header.Add("X-Domain-Id", fmt.Sprintf("%d", dom.ID))
	proxyReq.Header.Add("X-Domain-Host", dom.HostPublic)
	proxyReq.Header.Add("X-Domain-Skin", dom.Skin)

	//Send the proxy request using the custom transport
	resp, err := s.customTransport.RoundTrip(proxyReq)
	if err != nil {
		log.Println("Proxy error", err)
		http.Error(w, "Proxy error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	pubURL := dom.SchemePublic + "://" + dom.HostPublic
	if dom.PortPublic != "" {
		pubURL += ":" + dom.PortPublic
	}

	// Copy the headers from the proxy response to the original response
	needReplaceDomain := false
	needReplaceCanonical := false
	for name, values := range resp.Header {
		for _, value := range values {
			if name == "X-Powered-By" {
				continue
			}
			if name == "Server" {
				continue
			}
			if name == "Content-Length" {
				continue
			}
			if name == "Location" {
				value = strings.ReplaceAll(value, "https://"+dom.HostPrivate, pubURL)
			}
			// need to modify html
			if name == "Content-Type" && (strings.HasPrefix(value, "text/html") || strings.HasPrefix(value, "application/xml") || strings.HasPrefix(value, "application/json") || strings.HasPrefix(value, "text/plain")) {
				needReplaceDomain = true
			}
			if name == "Content-Type" && (strings.HasPrefix(value, "text/html")) {
				needReplaceCanonical = true
			}
			//log.Println("response header:", name, value)
			w.Header().Add(name, value)
		}
	}

	if needReplaceDomain && !forbiddenReplaceDomain {
		body, _ := io.ReadAll(resp.Body)

		//log.Printf("%s %d R\n", path, len(body))

		pubURLHost := strings.ReplaceAll(pubURL, "https://", "")
		//body = bytes.ReplaceAll(body, []byte("//"+dom.HostPrivate), []byte(pubURL))
		// body = bytes.ReplaceAll(body, []byte("http://"+dom.HostPrivate), []byte(pubURL))

		// sometimes we have urls in public sites to admin domain, replace them too!
		body = bytes.ReplaceAll(body, []byte("odminko."+dom.HostPrivate), []byte(pubURLHost))
		body = bytes.ReplaceAll(body, []byte("odminko.printhouse.casa"), []byte(pubURLHost))

		body = bytes.ReplaceAll(body, []byte(dom.HostPrivate), []byte(pubURLHost))

		//body = bytes.ReplaceAll(body, []byte("https://"+dom.HostPrivate), []byte(pubURL))

		// remove S3 domain for images
		body = bytes.ReplaceAll(body, []byte(dom.ServiceImager), []byte(""))

		// cache breaker for all images
		//body = bytes.ReplaceAll(body, []byte(".jpg\""), []byte(".jpg?v=1\""))

		if needReplaceCanonical {
			//log.Println("replace canonical")
			// <link rel="canonical" href="http://qwe/rwrrfewr/page/2/">
			// <link rel="canonical" href="http://qwe/rwrrfewr/">
			re := regexp.MustCompile(`<link rel="canonical" href="(.*)\/page\/[0-9]+">`)
			body = re.ReplaceAll(body, []byte(`<link rel="canonical" href="${1}">`))
		}

		w.Header().Set("X-Proxy-Mode", "modified")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))

		log.Printf("%s (%s) %s %d R\n", r.Method, host, targetURL, len(body))

		w.Header().Add("X-Proxy-tm", fmt.Sprintf("%d", time.Since(start).Milliseconds()))
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	//log.Printf("%s\n", path)

	// это тупо конечно вычитывать ответ только чтобы узнать его длину но апач не передает Content-Length
	body, _ := io.ReadAll(resp.Body)
	log.Printf("%s (%s) %s %d D\n", r.Method, host, targetURL, len(body))
	w.Header().Set("X-Proxy-Mode", "direct")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.Header().Add("X-Proxy-tm", fmt.Sprintf("%d", time.Since(start).Milliseconds()))
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
	// io.Copy(w, resp.Body)
}

func isHopHeader(header string) bool {
	for _, h := range hopHeaders {
		if header == h {
			return true
		}
	}
	return false
}
