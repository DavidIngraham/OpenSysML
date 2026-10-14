package resolve

import (
	"sync"

	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
)

// SharedCollisions is the import-collision memo (importedCollisions) the
// documents of one batch share: each is analyzed in a resolver of its own,
// and without it each walked the chain of namespaces its imports re-export
// from, so a batch did quadratic work in the chain's depth. An entry is what
// one resolver settled, with what it read of the index to settle it, which a
// resolver taking the entry reads again so its document's record still
// depends on all of it. The index is fixed for a batch, so what one namespace's
// imports bring is the same from every document.
type SharedCollisions struct {
	mu      sync.RWMutex
	entries map[*symbols.Scope]*sharedCollision
	below   map[*symbols.Scope]*sharedBelow
}

type sharedCollision struct {
	c     *importCollisions
	reads *readCapture
}

// sharedBelow is one settled hidesBelow answer with what settling it read.
type sharedBelow struct {
	hides bool
	reads *readCapture
}

// NewSharedCollisions returns an empty memo for one batch to share.
func NewSharedCollisions() *SharedCollisions {
	return &SharedCollisions{
		entries: map[*symbols.Scope]*sharedCollision{},
		below:   map[*symbols.Scope]*sharedBelow{},
	}
}

func (s *SharedCollisions) getBelow(scope *symbols.Scope) (*sharedBelow, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.below[scope]
	return e, ok
}

func (s *SharedCollisions) putBelow(scope *symbols.Scope, hides bool, reads *readCapture) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, had := s.below[scope]; !had {
		s.below[scope] = &sharedBelow{hides: hides, reads: reads}
	}
}

func (s *SharedCollisions) get(scope *symbols.Scope) (*sharedCollision, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[scope]
	return e, ok
}

func (s *SharedCollisions) put(scope *symbols.Scope, c *importCollisions, reads *readCapture) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, had := s.entries[scope]; !had {
		s.entries[scope] = &sharedCollision{c: c, reads: reads}
	}
}

// ShareCollisions has r read and settle import collisions through shared.
func (r *Resolver) ShareCollisions(shared *SharedCollisions) {
	if r != nil {
		r.shared = shared
	}
}

// readCapture gathers what the index answered while one computation ran, so
// a shared result can be read again by whoever takes it (see Resolver.Read*).
type readCapture struct {
	names, namespaces, docs, segments []string
	all, spellings                    bool
}

func (c *readCapture) replay(r *Resolver) {
	if c == nil {
		return
	}
	for _, fqn := range c.names {
		r.ReadName(fqn)
	}
	for _, fqn := range c.namespaces {
		r.ReadNamespace(fqn)
	}
	for _, name := range c.docs {
		r.ReadDocument(name)
	}
	for _, name := range c.segments {
		r.ReadSegment(name)
	}
	if c.all {
		r.ReadAllNames()
	}
	if c.spellings {
		r.ReadSpellings()
	}
}

// capturing runs compute with a fresh capture on the stack and returns it; an
// enclosing capture sees the same reads, as its computation made them too.
func (r *Resolver) capturing(compute func()) *readCapture {
	c := &readCapture{}
	r.captures = append(r.captures, c)
	defer func() { r.captures = r.captures[:len(r.captures)-1] }()
	compute()
	return c
}
