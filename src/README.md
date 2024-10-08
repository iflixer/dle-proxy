
alternative request
	// client := &http.Client{}
	// resp, err := client.Do(proxyReq)
	// if err != nil {
	// 	http.Error(w, "Server Error", http.StatusInternalServerError)
	// 	log.Println("ServeHTTP:", err)
	// }
	// defer resp.Body.Close()

debug
	// req, _ := http.NewRequest("GET", "http://1.2.3.4", nil)
	// req.Host = "www.example.org"
	// dump, _ := httputil.DumpRequestOut(req, false)
	// fmt.Printf("%s", dump)



	{"http":{"routers":{"cis-proxy":{"service":"cis-proxy","rule":"Host(`cis-proxy`)"},"traefik":{"middlewares":["traefik-auth"],"service":"api@internal","rule":"Host(`traefik.testme.cloud`)"},"whoami":{"service":"whoami","rule":"Host(`whoami.testme.cloud`,`whoami2.testme.cloud`)"}},"services":{"cis-proxy":{"loadBalancer":{"servers":[{"url":"http://10.0.3.101:80"},{"url":"http://10.0.3.100:80"},{"url":"http://10.0.3.102:80"},{"url":"http://10.0.3.99:80"},{"url":"http://10.0.3.103:80"}],"passHostHeader":true}},"traefik":{"loadBalancer":{"servers":[{"url":"http://10.0.3.98:8080"},{"url":"http://10.0.3.93:8080"},{"url":"http://10.0.3.95:8080"},{"url":"http://10.0.3.97:8080"},{"url":"http://10.0.3.94:8080"},{"url":"http://10.0.3.96:8080"}],"passHostHeader":true}},"whoami":{"loadBalancer":{"servers":[{"url":"http://10.0.3.41:8082"},{"url":"http://10.0.3.38:8082"},{"url":"http://10.0.3.40:8082"},{"url":"http://10.0.3.42:8082"},{"url":"http://10.0.3.39:8082"}],"passHostHeader":true}}},"middlewares":{"traefik-auth":{"basicAuth":{"users":["admin:$apr1$CtjaWI28$uabbOfKyM/lfXd0k4ib821"]}}}},"tcp":{},"udp":{}}