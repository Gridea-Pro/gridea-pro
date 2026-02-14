package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"sync"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	nanoIDAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	nanoIDLength   = 6
)

type LinkService struct {
	repo domain.LinkRepository
	mu   sync.RWMutex
}

func NewLinkService(repo domain.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

func (s *LinkService) LoadLinks(ctx context.Context) ([]domain.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetAll(ctx)
}

func (s *LinkService) SaveLinks(ctx context.Context, links []domain.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.SaveAll(ctx, links)
}

func (s *LinkService) SaveLink(ctx context.Context, link domain.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	links, err := s.repo.GetAll(ctx)
	if err != nil {
		links = []domain.Link{}
	}

	found := false
	for i := range links {
		if links[i].ID == link.ID {
			links[i] = link
			found = true
			break
		}
	}

	if !found {
		if link.ID == "" {
			// fallback if ID missing for new link (should check)
			id, err := gonanoid.Generate(nanoIDAlphabet, nanoIDLength)
			if err != nil {
				return fmt.Errorf("failed to generate link ID: %w", err)
			}
			link.ID = id
		}
		links = append(links, link)
	}

	return s.repo.SaveAll(ctx, links)
}

func (s *LinkService) DeleteLink(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	links, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	filtered := make([]domain.Link, 0, len(links))
	found := false
	for _, link := range links {
		if link.ID != id {
			filtered = append(filtered, link)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("link not found")
	}

	return s.repo.SaveAll(ctx, filtered)
}

// FixMissingIDs checks and repairs missing IDs in links.
// Returns true if any changes were made.
func (s *LinkService) FixMissingIDs(ctx context.Context) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	links, err := s.repo.GetAll(ctx)
	if err != nil {
		return false, err
	}

	hasMissingID := false
	for i := range links {
		if links[i].ID == "" {
			id, err := gonanoid.Generate(nanoIDAlphabet, nanoIDLength)
			if err == nil {
				links[i].ID = id
				hasMissingID = true
				fmt.Printf("Service Patched missing ID for link: %s -> %s\n", links[i].Name, id)
			}
		}
	}

	if hasMissingID {
		if err := s.repo.SaveAll(ctx, links); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}
