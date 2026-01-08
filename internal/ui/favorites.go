// Package ui provides the terminal user interface for Repo-lyzer.
// This file implements the favorites/bookmarks functionality for repositories.
package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Favorite represents a bookmarked repository with usage tracking.
type Favorite struct {
	RepoName string    `json:"repo_name"` // Full repository name (owner/repo)
	AddedAt  time.Time `json:"added_at"`  // When the favorite was added
	LastUsed time.Time `json:"last_used"` // Last time the repo was analyzed
	UseCount int       `json:"use_count"` // Number of times analyzed
	Notes    string    `json:"notes"`     // Optional user notes
}

// Favorites manages a collection of favorite repositories.
type Favorites struct {
	Items []Favorite `json:"items"`
}

// getFavoritesPath returns the path to the favorites JSON file.
func getFavoritesPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".repo-lyzer")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "favorites.json"), nil
}

// LoadFavorites loads favorites from the JSON file.
// Returns an empty Favorites struct if the file doesn't exist.
func LoadFavorites() (*Favorites, error) {
	path, err := getFavoritesPath()
	if err != nil {
		return &Favorites{Items: []Favorite{}}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Favorites{Items: []Favorite{}}, nil
		}
		return &Favorites{Items: []Favorite{}}, err
	}

	var favs Favorites
	if err := json.Unmarshal(data, &favs); err != nil {
		return &Favorites{Items: []Favorite{}}, err
	}

	return &favs, nil
}

// Save persists the favorites to the JSON file.
func (f *Favorites) Save() error {
	path, err := getFavoritesPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Add adds a repository to favorites or updates it if already exists.
func (f *Favorites) Add(repoName string) {
	now := time.Now()
	
	// Check if already exists
	for i, fav := range f.Items {
		if fav.RepoName == repoName {
			f.Items[i].UseCount++
			f.Items[i].LastUsed = now
			return
		}
	}

	// Add new favorite
	f.Items = append(f.Items, Favorite{
		RepoName: repoName,
		AddedAt:  now,
		LastUsed: now,
		UseCount: 1,
	})
}

// Remove removes a repository from favorites.
func (f *Favorites) Remove(repoName string) {
	for i, fav := range f.Items {
		if fav.RepoName == repoName {
			f.Items = append(f.Items[:i], f.Items[i+1:]...)
			return
		}
	}
}

// IsFavorite checks if a repository is in favorites.
func (f *Favorites) IsFavorite(repoName string) bool {
	for _, fav := range f.Items {
		if fav.RepoName == repoName {
			return true
		}
	}
	return false
}

// UpdateUsage updates the usage statistics for a favorite.
func (f *Favorites) UpdateUsage(repoName string) {
	for i, fav := range f.Items {
		if fav.RepoName == repoName {
			f.Items[i].UseCount++
			f.Items[i].LastUsed = time.Now()
			return
		}
	}
}

// GetTopFavorites returns the top N favorites sorted by use count.
func (f *Favorites) GetTopFavorites(n int) []Favorite {
	if n <= 0 {
		return []Favorite{}
	}

	// Make a copy to sort
	sorted := make([]Favorite, len(f.Items))
	copy(sorted, f.Items)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].UseCount > sorted[j].UseCount
	})

	if n > len(sorted) {
		return sorted
	}
	return sorted[:n]
}

// Clear removes all favorites.
func (f *Favorites) Clear() {
	f.Items = []Favorite{}
}
