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

var bidPool = sync.Pool{
	New: func() any {
		var val BidType = BidType{byte(0), byte(0)}
		return &val
	},
}

type Service struct {
	Data      *Data
	StatsLock *sync.RWMutex
	Stats     map[BidType]int
}

func NewService() *Service {
	return &Service{
		Data:      NewData(),
		StatsLock: new(sync.RWMutex),
		Stats:     make(map[BidType]int),
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
	b := bidPool.Get().(*BidType)
	b[0] = Chars[rand.Intn(len(Chars))]
	b[1] = Chars[rand.Intn(len(Chars))]
	randomIndex := rand.Intn(NumberOfBids)

	s.Data.Put(randomIndex, b)
}

// TODO: check if bid is unique for array
func (s *Service) Populate() {
	for i := 0; i < NumberOfBids; i++ {
		b := bidPool.Get().(*BidType)
		b[0] = Chars[rand.Intn(len(Chars))]
		b[1] = Chars[rand.Intn(len(Chars))]
		s.Data.Put(i, b)
	}
}

func (s *Service) RequestHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/plain")
	ctx.SetStatusCode(fasthttp.StatusOK)
	randomIndex := rand.Intn(NumberOfBids)
	value := s.Data.Get(randomIndex)
	var slice []byte = value[:]
	ctx.SetBody(slice)
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

func (s *Service) UpdateStats(k *BidType) {
	key := *k
	s.StatsLock.Lock()
	defer s.StatsLock.Unlock()

	s.Stats[key] = s.Stats[key] + 1
}

func (s *Service) GetStats() map[string]int {
	result := make(map[string]int)
	s.StatsLock.RLock()
	for key, value := range s.Stats {
		var slice []byte = key[:]
		result[string(slice)] = value
	}
	defer s.StatsLock.RUnlock()

	return result
}
