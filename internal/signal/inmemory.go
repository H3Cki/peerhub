package sig

import (
	"sync"

	"github.com/H3Cki/peerhub"
)

type InMemoryService struct {
	mu      sync.Mutex
	offers  map[string]peerhub.Offer
	answers map[string]peerhub.Answer
}

func NewInMemoryService() *InMemoryService {
	return &InMemoryService{
		mu:      sync.Mutex{},
		offers:  map[string]peerhub.Offer{},
		answers: map[string]peerhub.Answer{},
	}
}

func (s *InMemoryService) CreateOffer(o peerhub.Offer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offers[o.ID] = o
	return nil
}

func (s *InMemoryService) GetOffer(offerID string) (peerhub.Offer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.offers[offerID]
	if !ok {
		return peerhub.Offer{}, peerhub.ErrOfferNotFound
	}
	return o, nil
}

func (s *InMemoryService) DeleteOffer(offerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.offers, offerID)
	return nil
}

func (s *InMemoryService) CreateAnswer(a peerhub.Answer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.answers[a.ID] = a
	return nil
}

func (s *InMemoryService) GetAnswer(answerID string) (peerhub.Answer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.answers[answerID]
	if !ok {
		return peerhub.Answer{}, peerhub.ErrAnswerNotFound
	}
	return a, nil
}

func (s *InMemoryService) DeleteAnswer(answerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.answers, answerID)
	return nil
}
