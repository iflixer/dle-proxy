package domainAlias

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
	domains      []*DomainAlias
}

type DomainAlias struct {
	DomainID int
	Host     string
}

func (c *DomainAlias) TableName() string {
	return "flix_domain_alias"
}

func (s *Service) GetDomain(host string) (domain DomainAlias, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, g := range s.domains {
		if g.Host == host {
			return *g, nil
		}
	}

	return domain, fmt.Errorf("host not found:%s", host)
}

func (s *Service) GetDomains() (domains []DomainAlias, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, g := range s.domains {
		domains = append(domains, *g)
	}

	return
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
	var dd []*DomainAlias
	if err = s.dbService.DB.Find(&dd).Error; err == nil {
		s.mu.Lock()
		s.domains = dd
		s.mu.Unlock()
	}
	return
}
