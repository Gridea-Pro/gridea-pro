package domain

import (
	"context"
	"errors"
	"strings"
)

// Tag 标签结构
// Added json tags for frontend compatibility.
type Tag struct {
	ID    string `json:"_id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Used  bool   `json:"used"`
	Color string `json:"color,omitempty"`
}

// Validate checking
func (t *Tag) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return errors.New("tag name is required")
	}
	if strings.TrimSpace(t.Slug) == "" {
		return errors.New("tag slug is required")
	}
	return nil
}

// TagRepository defines interface
type TagRepository interface {
	List(ctx context.Context) ([]Tag, error)
	// Standard CRUD additions if needed
	Create(ctx context.Context, tag *Tag) error
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id string) error
	SaveAll(ctx context.Context, tags []Tag) error
}

// GetID implements Identifiable interface
func (t Tag) GetID() string {
	return t.ID
}
