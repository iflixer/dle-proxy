package server

import (
	"dle-proxy/database"
	"dle-proxy/database/domain"
	"dle-proxy/database/domainAlias"
	"dle-proxy/database/domainFile"
	"dle-proxy/database/flixPost"
	"fmt"
	"log"
	"net/http"
)

type Service struct {
	port               string
	server             http.Server
	dbService          *database.Service
	domainService      *domain.Service
	domainAliasService *domainAlias.Service
	fileService        *domainFile.Service
	flixPostService    *flixPost.Service
	customTransport    http.RoundTripper
}

func (s *Service) Run() {
	addr := fmt.Sprintf(":%s", s.port)
	log.Println("Starting proxy server on", addr)
	err := s.server.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting proxy server: ", err)
	}
}

func NewService(port string, dbService *database.Service, domainService *domain.Service, domainAliasService *domainAlias.Service, fileService *domainFile.Service, flixPostService *flixPost.Service) (s *Service, err error) {

	s = &Service{
		port:               port,
		dbService:          dbService,
		domainService:      domainService,
		domainAliasService: domainAliasService,
		fileService:        fileService,
		flixPostService:    flixPostService,
		customTransport:    http.DefaultTransport,
	}
	s.server = http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: http.HandlerFunc(s.Proxy),
	}

	return
}
