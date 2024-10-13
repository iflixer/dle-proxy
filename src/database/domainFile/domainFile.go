package domainFile

import (
	"dle-proxy/database"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type Service struct {
	mu           sync.RWMutex
	dbService    *database.Service
	updatePeriod time.Duration
	files        []*DomainFile
}

type DomainFile struct {
	ID          int
	DomainId    int
	Path        string
	ContentType string
	Body        string
}

func (c *DomainFile) TableName() string {
	return "flix_domain_files"
}

func (s *Service) GetFile(domainId int, path string) (domain *DomainFile, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path = strings.Trim(path, "/")

	log.Println("search file by ", domainId, path)
	for _, g := range s.files {
		if g.DomainId == domainId && g.Path == path {
			return g, nil
		}
	}

	return nil, fmt.Errorf("file not found:%d %s", domainId, path)
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
	var dd []*DomainFile
	if err = s.dbService.DB.Find(&dd).Error; err == nil {
		s.mu.Lock()
		s.files = dd
		s.mu.Unlock()
	}
	return
}
