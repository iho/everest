package everest

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	Chars string = "abcdefghijklmnopqrstuvwxyz"
)

type Service struct {
	Data      *Data
	StatsLock *sync.RWMutex
	Stats     map[string]int
}

func NewService() *Service {
	return &Service{
		Data:      NewData(),
		StatsLock: new(sync.RWMutex),
		Stats:     make(map[string]int),
	}
}

func (s *Service) Ticker(d time.Duration) {
	for {
		time.Sleep(d)
		s.Tick()
	}
}

// TODO: check if bid is unique for array
func (s *Service) Tick() {
	value := string(Chars[rand.Intn(len(Chars))]) + string(Chars[rand.Intn(len(Chars))])
	randomIndex := rand.Intn(NumberOfBids)

	s.Data.Put(randomIndex, value)
}

// TODO: check if bid is unique for array
func (s *Service) Populate() {
	for i := 0; i < NumberOfBids; i++ {
		value := string(Chars[rand.Intn(len(Chars))]) + string(Chars[rand.Intn(len(Chars))])
		s.Data.Put(i, value)
	}
}

func (s *Service) RequestHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/plain")
	ctx.SetStatusCode(fasthttp.StatusOK)
	randomIndex := rand.Intn(NumberOfBids)
	value := s.Data.Get(randomIndex)
	ctx.SetBody([]byte(value))
	go s.UpdateStats(value) // TODO: benchmark and improve
}

func (s *Service) AdminHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/plain")
	ctx.SetStatusCode(fasthttp.StatusOK)

	// TODO: make it faster if needed
	response := ""
	// for key, value := range s.GetStats() {
	// 	if value == 0 {
	// 		continue
	// 	}
	// 	response += fmt.Sprintf("%s - %d\n", key, value)
	// }

	m := s.GetStats()
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if m[keys[i]] == m[keys[j]] {
			return keys[i] < keys[j]
		}
		return m[keys[i]] > m[keys[j]]
	})
	for _, k := range keys {
		response += fmt.Sprintf("%s - %d\n", k, m[k])
	}

	ctx.SetBody([]byte(response))
}

func (s *Service) UpdateStats(key string) {
	s.StatsLock.Lock()
	defer s.StatsLock.Unlock()

	s.Stats[key] = s.Stats[key] + 1
}

func (s *Service) GetStats() map[string]int {
	result := make(map[string]int)
	s.StatsLock.RLock()
	for key, value := range s.Stats {
		result[key] = value
	}
	defer s.StatsLock.RUnlock()

	return result
}
