package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/WAT36/shorty/internal/shortener"
)

type Mapping struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Clicks    int       `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
}

type FileStore struct {
	mu     sync.RWMutex
	path   string
	items  map[string]Mapping
	loaded bool
}

func NewFileStore(path string) (*FileStore, error) {
	if path == "" {
		return nil, errors.New("empty store path")
	}
	return &FileStore{
		path:  path,
		items: make(map[string]Mapping),
	}, nil
}

func (s *FileStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		// 初回などファイルがない場合
		return err
	}
	var items []Mapping
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	s.items = make(map[string]Mapping, len(items))
	for _, m := range items {
		s.items[m.Code] = m
	}
	s.loaded = true
	return nil
}

func (s *FileStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	items := make([]Mapping, 0, len(s.items))
	for _, m := range s.items {
		items = append(items, m)
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *FileStore) Create(rawURL, custom string) (string, error) {
	if rawURL == "" {
		return "", errors.New("url required")
	}
	code := custom
	if code != "" {
		if err := shortener.ValidateCustom(code); err != nil {
			return "", err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// カスタム重複チェック
	if code != "" {
		if _, ok := s.items[code]; ok {
			return "", errors.New("custom code already exists")
		}
	} else {
		// ランダム生成（重複しないまで）
		for {
			c, err := shortener.RandomCode()
			if err != nil {
				return "", err
			}
			if _, ok := s.items[c]; !ok {
				code = c
				break
			}
		}
	}

	m := Mapping{
		Code:      code,
		URL:       rawURL,
		Clicks:    0,
		CreatedAt: time.Now(),
	}
	s.items[code] = m
	if err := s.Save(); err != nil {
		return "", fmt.Errorf("save failed: %w", err)
	}
	return code, nil
}

func (s *FileStore) Get(code string) (Mapping, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.items[code]
	return m, ok
}

func (s *FileStore) List() []Mapping {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Mapping, 0, len(s.items))
	for _, m := range s.items {
		out = append(out, m)
	}
	return out
}

func (s *FileStore) Delete(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[code]; !ok {
		return errors.New("not found")
	}
	delete(s.items, code)
	return s.Save()
}

func (s *FileStore) Increment(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.items[code]
	if !ok {
		return errors.New("not found")
	}
	m.Clicks++
	s.items[code] = m
	return s.Save()
}
