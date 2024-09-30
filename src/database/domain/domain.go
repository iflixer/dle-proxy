package domain

import (
	"dle-proxy/database"
	"fmt"
	"log"
	"sync"
	"time"
)

type Service struct {
	mu           sync.RWMutex
	dbService    *database.Service
	updatePeriod time.Duration
	domains      []*Domain
}

type Domain struct {
	ID             int
	Title          string
	HostPublic     string
	HostPrivate    string
	Skin           string
	ServiceDle     string
	ServiceImager  string
	ServiceSitemap string
	NewsNumber     int
	PortPublic     string
	SchemePublic   string
}

func (c *Domain) TableName() string {
	return "flix_domain"
}

func (s *Service) GetDomain(host string) (domain *Domain, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, g := range s.domains {
		if g.HostPublic == host {
			return g, nil
		}
	}

	return nil, fmt.Errorf("host not found:%s", host)
}

func NewService(dbService *database.Service, updatePeriod int) (s *Service, err error) {

	s = &Service{
		dbService:    dbService,
		updatePeriod: time.Duration(updatePeriod),
	}

	err = s.loadData()

	go s.loadWorker()

	return
}

func (s *Service) loadWorker() {
	for {
		time.Sleep(time.Second * s.updatePeriod)
		if err := s.loadData(); err != nil {
			log.Println(err)
		}
	}
}

func (s *Service) loadData() (err error) {
	var dd []*Domain
	if err = s.dbService.DB.Find(&dd).Error; err == nil {
		s.mu.Lock()
		s.domains = dd
		s.mu.Unlock()
	}
	return
}
