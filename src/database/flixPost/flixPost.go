package flixPost

import (
	"dle-proxy/database"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type Service struct {
	mu           sync.RWMutex
	dbService    *database.Service
	updatePeriod time.Duration
	flixPosts    map[int]FlixPost
}

type FlixPost struct {
	ID       int
	DomainID int
	PostID   int
	AltName  string
	Approve  bool
	Redirect int
}

func (c *FlixPost) TableName() string {
	return "flix_post"
}

func (s *Service) GetPost(domainID int, u string) (post FlixPost, altName string, err error) {
	re := regexp.MustCompile(`\/([0-9]+)\-(.*)\.html`)
	parts := re.FindStringSubmatch(u)
	if len(parts) < 3 {
		return post, "", fmt.Errorf("cant regexp ID and altName in url: %s", u)
	}
	postID, err := strconv.Atoi(parts[1])
	if (err != nil) || (postID == 0) {
		return post, "", fmt.Errorf("cant find ID in url: %s", u)
	}
	altName = parts[2]

	postKey := s.generateFlixPostKey(domainID, postID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	if post, ok := s.flixPosts[postKey]; ok {
		return post, altName, nil
	}

	return post, "", fmt.Errorf("flixPost not found for url: %s", u)

}

func NewService(dbService *database.Service, updatePeriod int) (s *Service, err error) {
	s = &Service{
		dbService:    dbService,
		updatePeriod: time.Duration(updatePeriod),
		flixPosts:    make(map[int]FlixPost),
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
	var rows []*FlixPost
	if err = s.dbService.DB.Find(&rows).Error; err == nil {
		s.mu.Lock()
		for _, row := range rows {
			id := s.generateFlixPostKey(row.DomainID, row.PostID)
			s.flixPosts[id] = *row
		}
		s.mu.Unlock()
	}
	return
}

func (s *Service) generateFlixPostKey(domainID, postID int) int {
	return domainID*10000 + postID
}
