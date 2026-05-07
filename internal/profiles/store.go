package profiles

import "errors"

type Store interface {
	List(activeID string) []Profile
	Get(id string) (Profile, bool)
	Save(activeID string) error
}

type InMemoryStore struct {
	profiles []Profile
	activeID string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{profiles: DefaultProfiles(), activeID: "default"}
}

func (s *InMemoryStore) List(activeID string) []Profile {
	if activeID == "" {
		activeID = s.activeID
	}
	result := make([]Profile, 0, len(s.profiles))
	for _, p := range s.profiles {
		cp := p
		cp.IsActive = p.ID == activeID
		result = append(result, cp)
	}
	return result
}

func (s *InMemoryStore) Get(id string) (Profile, bool) {
	for _, p := range s.profiles {
		if p.ID == id {
			return p, true
		}
	}
	return Profile{}, false
}

func (s *InMemoryStore) Save(activeID string) error {
	if _, ok := s.Get(activeID); !ok {
		return errors.New("profile not found")
	}
	s.activeID = activeID
	return nil
}
