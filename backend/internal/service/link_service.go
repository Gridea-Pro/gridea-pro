package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"sync"
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
	return s.repo.List(ctx)
}

func (s *LinkService) SaveLinks(ctx context.Context, links []domain.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.SaveAll(ctx, links)
}

func (s *LinkService) CreateLink(ctx context.Context, link domain.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	links, err := s.repo.List(ctx)
	if err != nil {
		fmt.Printf("LinkService: Failed to list links: %v\n", err)
		return err
	}

	links = append(links, link)

	if err := s.repo.SaveAll(ctx, links); err != nil {
		fmt.Printf("LinkService: Failed to save links: %v\n", err)
		return err
	}
	return nil
}

func (s *LinkService) UpdateLink(ctx context.Context, link domain.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	links, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	found := false
	for i, l := range links {
		if l.ID == link.ID {
			links[i] = link
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("link not found")
	}

	return s.repo.SaveAll(ctx, links)
}

func (s *LinkService) DeleteLink(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	links, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	var newLinks []domain.Link
	found := false // Added found flag for consistency with other methods
	for _, l := range links {
		if l.ID != id {
			newLinks = append(newLinks, l)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("link not found")
	}

	return s.repo.SaveAll(ctx, newLinks)
}
