package cron

import (
	"sync"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron    *cron.Cron
	entries map[uint]cron.EntryID
	mu      sync.Mutex
}

func NewScheduler() *Scheduler {
	c := cron.New(cron.WithSeconds())
	c.Start()
	return &Scheduler{cron: c, entries: make(map[uint]cron.EntryID)}
}

func (s *Scheduler) Add(id uint, expr string, fn func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if eid, ok := s.entries[id]; ok {
		s.cron.Remove(eid)
	}
	eid, err := s.cron.AddFunc(expr, fn)
	if err != nil {
		return err
	}
	s.entries[id] = eid
	return nil
}

func (s *Scheduler) Remove(id uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if eid, ok := s.entries[id]; ok {
		s.cron.Remove(eid)
		delete(s.entries, id)
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := s.cron.Stop()
	<-ctx.Done()
}
