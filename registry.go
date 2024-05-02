package sdphub

import (
	"fmt"
	"slices"
	"sync"

	"golang.org/x/exp/maps"
)

type Registry struct {
	mu        sync.Mutex
	answerers map[string]Answerer
	offerers  map[string][]Offerer

	agreements map[string]Agreement
}

func NewRegistry() *Registry {
	return &Registry{
		mu:         sync.Mutex{},
		answerers:  map[string]Answerer{},
		offerers:   map[string][]Offerer{},
		agreements: map[string]Agreement{},
	}
}

func (r *Registry) CreateAnswerer(answerer Answerer) (*Answerer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var replaced *Answerer
	reg, ok := r.answerers[answerer.Name]
	if ok && !reg.ManagementKeyMatches(answerer.ManagementKey) {
		return nil, fmt.Errorf("answerer with this name already exists")
	}

	if ok {
		replaced = &reg
	}

	r.answerers[answerer.Name] = answerer
	return replaced, nil
}

func (r *Registry) Answerers() []Answerer {
	r.mu.Lock()
	defer r.mu.Unlock()
	return maps.Values(r.answerers)
}

func (r *Registry) Answerer(registrantName string) (Answerer, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	registrant, ok := r.answerers[registrantName]
	return registrant, ok
}

func (r *Registry) Agreement(agrID string) (Agreement, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.agreements[agrID]
	return c, ok
}

func (r *Registry) CreateAgreement(agr Agreement) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agreements[agr.ID] = agr
}

func (r *Registry) DeleteAgreement(agrID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agreements, agrID)
}

func (r *Registry) CreateOfferer(offerer Offerer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.offerers[offerer.Name] = append(r.offerers[offerer.Name], offerer)
	return nil
}

func (r *Registry) Offerer(answererName, offererName string) (Offerer, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	offerers, ok := r.offerers[answererName]
	if !ok {
		return Offerer{}, false
	}

	idx := slices.IndexFunc(offerers, func(o Offerer) bool { return o.Name == offererName })
	if idx == -1 {
		return Offerer{}, false
	}

	return offerers[idx], true
}

func (r *Registry) Offerers(answererName string) []Offerer {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.offerers[answererName]
}
