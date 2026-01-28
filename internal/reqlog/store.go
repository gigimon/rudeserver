package reqlog

import (
	"sync"
	"time"
)

type Entry struct {
	ID        int64
	Timestamp time.Time
	Method    string
	Path      string
	Query     string
	Protocol  string
	ClientIP  string
	Status    int
	Duration  int64

	ReqHeaders map[string][]string
	ResHeaders map[string][]string

	ReqBody       []byte
	ResBody       []byte
	ReqTruncated  bool
	ResTruncated  bool
	ReqSize       int64
	ResSize       int64
	ContentType  string
	ReqBodyIsUTF bool
	ResBodyIsUTF bool
	ReqBodyB64   string
	ResBodyB64   string
	ReqError     string
	ResError     string
	UserAgent    string
}

type Store struct {
	mu      sync.Mutex
	entries []Entry
	nextID  int64
	max     int
}

func NewStore(max int) *Store {
	if max <= 0 {
		max = 100
	}
	return &Store{
		entries: make([]Entry, 0, max),
		nextID:  1,
		max:     max,
	}
}

func (s *Store) Add(entry Entry) Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry.ID = s.nextID
	s.nextID++
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	s.entries = append([]Entry{entry}, s.entries...)
	if len(s.entries) > s.max {
		s.entries = s.entries[:s.max]
	}
	return entry
}

func (s *Store) List() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

func (s *Store) Find(id int64) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range s.entries {
		if entry.ID == id {
			return entry, true
		}
	}
	return Entry{}, false
}
