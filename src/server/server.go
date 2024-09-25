package server

import (
	"dle-proxy/database"
	"dle-proxy/database/domain"
	"fmt"
	"log"
	"net/http"
)

type Service struct {
	port            string
	server          http.Server
	dbService       *database.Service
	domainService   *domain.Service
	customTransport http.RoundTripper
}

func (s *Service) Run() {
	addr := fmt.Sprintf(":%s", s.port)
	log.Println("Starting proxy server on", addr)
	err := s.server.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting proxy server: ", err)
	}
}

func NewService(port string, dbService *database.Service, domainService *domain.Service) (s *Service, err error) {

	s = &Service{
		port:            port,
		dbService:       dbService,
		domainService:   domainService,
		customTransport: http.DefaultTransport,
	}
	s.server = http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: http.HandlerFunc(s.Proxy),
	}

	return
}
