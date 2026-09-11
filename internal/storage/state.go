package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
)

type State struct {
	Documents map[string]string `json:"documents"`
}

func NewState() *State {
	return &State{
		Documents: make(map[string]string),
	}
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return NewState(), nil
	}

	if err != nil {
		return nil, err
	}

	state := NewState()

	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}

	return state, nil
}

func (s *State) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (s *State) HasChanged(url string, data []byte) bool {
	hash := Hash(data)

	oldHash, exists := s.Documents[url]

	return !exists || oldHash != hash
}

func (s *State) MarkProcessed(url string, data []byte) {
	s.Documents[url] = Hash(data)
}
