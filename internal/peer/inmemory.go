package peer

import (
	"sync"

	"github.com/H3Cki/peerhub"
	"golang.org/x/exp/maps"
)

type InMemoryService struct {
	mu  sync.Mutex
	ops map[string]peerhub.OfferingPeer
	aps map[string]peerhub.AnsweringPeer
}

func NewInMemoryService() *InMemoryService {
	return &InMemoryService{
		mu:  sync.Mutex{},
		ops: map[string]peerhub.OfferingPeer{},
		aps: map[string]peerhub.AnsweringPeer{},
	}
}

func (s *InMemoryService) GetAnsweringPeers() ([]peerhub.AnsweringPeer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return maps.Values(s.aps), nil
}

func (s *InMemoryService) CreateAnsweringPeer(ap peerhub.AnsweringPeer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aps[ap.Name] = ap
	return nil
}

func (s *InMemoryService) UpdateAnsweringPeer(ap peerhub.AnsweringPeer) error {
	return s.CreateAnsweringPeer(ap)
}

func (s *InMemoryService) GetAnsweringPeer(name string) (peerhub.AnsweringPeer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ap, ok := s.aps[name]
	if !ok {
		return peerhub.AnsweringPeer{}, peerhub.ErrAnsweringPeerNotFound
	}
	return ap, nil
}

func (s *InMemoryService) DeleteAnsweringPeer(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.aps, name)
	return nil
}

func (s *InMemoryService) CreateOfferingPeer(op peerhub.OfferingPeer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ops[op.Name] = op
	return nil
}

func (s *InMemoryService) UpdateOfferingPeer(op peerhub.OfferingPeer) error {
	return s.CreateOfferingPeer(op)
}

func (s *InMemoryService) GetOfferingPeer(name string) (peerhub.OfferingPeer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.ops[name]
	if !ok {
		return peerhub.OfferingPeer{}, peerhub.ErrOfferingPeerNotFound
	}
	return op, nil
}

func (s *InMemoryService) GetOfferingPeersByTarget(name string) ([]peerhub.OfferingPeer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ops := []peerhub.OfferingPeer{}
	for _, op := range s.ops {
		if op.TargetName == name {
			ops = append(ops, op)
		}
	}
	return ops, nil
}

func (s *InMemoryService) DeleteOfferingPeer(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.ops, name)
	return nil
}
