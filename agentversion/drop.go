package agentversion

import (
	"errors"
	"os"
)

// Drop removes one version's manifest record before deleting the blob.
// A failure after metadata commit leaves only an unreachable private blob,
// never an index pointing to a deleted historical version.
func (s *Store) Drop(id string) error {
	if !validKey(id) {
		return ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next := make([]Version, 0, len(s.data.Versions))
	found := false
	for _, v := range s.data.Versions {
		if v.ID == id {
			found = true
			continue
		}
		next = append(next, v)
	}
	if !found {
		return ErrNotFound
	}
	if err := s.persist(next); err != nil {
		return err
	}
	s.data.Versions = next
	if err := os.Remove(s.blobPath(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
