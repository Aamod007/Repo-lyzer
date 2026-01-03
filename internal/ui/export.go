package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ExportData is the structure for JSON export with additional metadata
type ExportData struct {
	ExportedAt    string         `json:"exported_at"`
	Repository    RepoExport     `json:"repository"`
	Metrics       MetricsExport  `json:"metrics"`
	Languages     map[string]int `json:"languages"`
	TopContributors []ContributorExport `json:"top_contributors"`
	CommitCount   int            `json:"commit_count_1y"`
}

type RepoExport struct {
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	Stars         int    `json:"stars"`
	Forks         int    `json:"forks"`
	OpenIssues    int    `json:"open_issues"`
	CreatedAt     string `json:"created_at"`
	LastPush      string `json:"last_push"`
	DefaultBranch string `json:"default_branch"`
	URL           string `json:"url"`
}

type MetricsExport struct {
	HealthScore   int    `json:"health_score"`
	BusFactor     int    `json:"bus_factor"`
	BusRisk       string `json:"bus_risk"`
	MaturityScore int    `json:"maturity_score"`
	MaturityLevel string `json:"maturity_level"`
}

type ContributorExport struct {
	Login   string `json:"login"`
	Commits int    `json:"commits"`
}

func ExportJSON(data AnalysisResult, filename string) error {
	// Build top contributors (max 10)
	var topContribs []ContributorExport
	maxContribs := 10
	if len(data.Contributors) < maxContribs {
		maxContribs = len(data.Contributors)
	}
	for i := 0; i < maxContribs; i++ {
		topContribs = append(topContribs, ContributorExport{
			Login:   data.Contributors[i].Login,
			Commits: data.Contributors[i].Commits,
		})
	}

	export := ExportData{
		ExportedAt: time.Now().Format(time.RFC3339),
		Repository: RepoExport{
			FullName:      data.Repo.FullName,
			Description:   data.Repo.Description,
			Stars:         data.Repo.Stars,
			Forks:         data.Repo.Forks,
			OpenIssues:    data.Repo.OpenIssues,
			CreatedAt:     data.Repo.CreatedAt.Format("2006-01-02"),
			LastPush:      data.Repo.PushedAt.Format("2006-01-02"),
			DefaultBranch: data.Repo.DefaultBranch,
			URL:           data.Repo.HTMLURL,
		},
		Metrics: MetricsExport{
			HealthScore:   data.HealthScore,
			BusFactor:     data.BusFactor,
			BusRisk:       data.BusRisk,
			MaturityScore: data.MaturityScore,
			MaturityLevel: data.MaturityLevel,
		},
		Languages:       data.Languages,
		TopContributors: topContribs,
		CommitCount:     len(data.Commits),
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(export)
}

func ExportMarkdown(data AnalysisResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	md := fmt.Sprintf("# Analysis for %s\n\n", data.Repo.FullName)
	md += fmt.Sprintf("*Exported: %s*\n\n", time.Now().Format("2006-01-02 15:04"))
	
	md += "## Repository Info\n"
	md += fmt.Sprintf("- **Stars:** %d\n", data.Repo.Stars)
	md += fmt.Sprintf("- **Forks:** %d\n", data.Repo.Forks)
	md += fmt.Sprintf("- **Open Issues:** %d\n", data.Repo.OpenIssues)
	md += fmt.Sprintf("- **Created:** %s\n", data.Repo.CreatedAt.Format("2006-01-02"))
	md += fmt.Sprintf("- **URL:** %s\n\n", data.Repo.HTMLURL)

	md += "## Metrics\n"
	md += fmt.Sprintf("- **Health Score:** %d/100\n", data.HealthScore)
	md += fmt.Sprintf("- **Bus Factor:** %d (%s)\n", data.BusFactor, data.BusRisk)
	md += fmt.Sprintf("- **Maturity:** %s (%d)\n", data.MaturityLevel, data.MaturityScore)
	md += fmt.Sprintf("- **Commits (1 year):** %d\n", len(data.Commits))
	md += fmt.Sprintf("- **Contributors:** %d\n\n", len(data.Contributors))

	md += "## Languages\n"
	total := 0
	for _, bytes := range data.Languages {
		total += bytes
	}
	for lang, bytes := range data.Languages {
		pct := float64(bytes) / float64(total) * 100
		md += fmt.Sprintf("- %s: %.1f%%\n", lang, pct)
	}
	md += "\n"

	md += "## Top Contributors\n"
	maxContribs := 10
	if len(data.Contributors) < maxContribs {
		maxContribs = len(data.Contributors)
	}
	for i := 0; i < maxContribs; i++ {
		c := data.Contributors[i]
		md += fmt.Sprintf("%d. %s (%d commits)\n", i+1, c.Login, c.Commits)
	}

	_, err = file.WriteString(md)
	return err
}
