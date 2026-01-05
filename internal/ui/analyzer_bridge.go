package ui

import (
	"github.com/agnivo988/Repo-lyzer/internal/analyzer"
	"github.com/agnivo988/Repo-lyzer/internal/github"
)

// AnalyzerDataBridge provides clean interface between analyzer logic and UI
type AnalyzerDataBridge struct {
	repo          *github.Repo
	commits       []github.Commit
	contributors  []github.Contributor
	languages     map[string]int
	healthScore   int
	busFactor     int
	busRisk       string
	maturityScore int
	maturityLevel string
	fileTree      *FileNode
	cache         map[string]interface{}
}

// NEW: Empty-state detection helper
func (b *AnalyzerDataBridge) IsEmpty() bool {
	return b.repo == nil ||
		(len(b.commits) == 0 &&
			len(b.contributors) == 0 &&
			len(b.languages) == 0)
}

// NewAnalyzerDataBridge creates a new data bridge with analyzer results
func NewAnalyzerDataBridge(result AnalysisResult) *AnalyzerDataBridge {
	bridge := &AnalyzerDataBridge{
		repo:          result.Repo,
		commits:       result.Commits,
		contributors:  result.Contributors,
		languages:     result.Languages,
		healthScore:   result.HealthScore,
		busFactor:     result.BusFactor,
		busRisk:       result.BusRisk,
		maturityScore: result.MaturityScore,
		maturityLevel: result.MaturityLevel,
 feat/empty-state-error-handling-58

		fileTree:      BuildFileTree(result),

	}

	// NEW: Avoid building tree if no commits exist
	if len(result.Commits) > 0 {
		bridge.fileTree = BuildFileTree(len(result.Commits), []string{})
	}

	return bridge
}


// GetHealthMetrics returns health-related metrics
func (b *AnalyzerDataBridge) GetHealthMetrics() map[string]interface{} {
	if b.IsEmpty() {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"health_score":   b.healthScore,
		"health_status":  b.getHealthStatus(),
		"bus_factor":     b.busFactor,
		"bus_risk":       b.busRisk,
		"maturity_level": b.maturityLevel,
		"maturity_score": b.maturityScore,
		"health_color":   b.getHealthColor(),
		"risk_color":     b.getRiskColor(),
	}
}

// GetRepositoryInfo returns repository metadata
func (b *AnalyzerDataBridge) GetRepositoryInfo() map[string]interface{} {
	if b.repo == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"name":           b.repo.FullName,
		"description":    b.repo.Description,
		"stars":          b.repo.Stars,
		"forks":          b.repo.Forks,
		"open_issues":    b.repo.OpenIssues,
		"default_branch": b.repo.DefaultBranch,
	}
}

// GetContributorMetrics returns contributor analysis
func (b *AnalyzerDataBridge) GetContributorMetrics() map[string]interface{} {
	if len(b.contributors) == 0 {
		return map[string]interface{}{
			"total_contributors": 0,
			"top_contributors":   []map[string]interface{}{},
			"contributor_count":  0,
			"diversity_score":    0,
		}
	}

	topContributors := b.getTopContributors(5)
	return map[string]interface{}{
		"total_contributors": len(b.contributors),
		"top_contributors":   topContributors,
		"contributor_count":  len(b.contributors),
		"diversity_score":    b.calculateDiversity(),
	}
}

// GetCommitMetrics returns commit-related metrics
func (b *AnalyzerDataBridge) GetCommitMetrics() map[string]interface{} {
	if len(b.commits) == 0 {
		return map[string]interface{}{
			"total_commits":    0,
			"commits_per_day":  map[string]int{},
			"recent_activity": map[string]int{},
			"commit_frequency": "No commits",
			"activity_trend":   "Unknown",
		}
	}

	commitActivity := analyzer.CommitsPerDay(b.commits)
	recentActivity := b.getRecentActivity()

	return map[string]interface{}{
		"total_commits":    len(b.commits),
		"commits_per_day":  commitActivity,
		"recent_activity":  recentActivity,
		"commit_frequency": b.calculateCommitFrequency(),
		"last_commit":      b.getLastCommitInfo(),
		"activity_trend":   b.calculateActivityTrend(),
	}
}

// GetLanguageMetrics returns programming language information
func (b *AnalyzerDataBridge) GetLanguageMetrics() map[string]interface{} {
	if len(b.languages) == 0 {
		return map[string]interface{}{
			"languages":          map[string]int{},
			"primary_language":   "Unknown",
			"language_count":     0,
			"language_diversity": 0,
		}
	}

	return map[string]interface{}{
		"languages":          b.languages,
		"primary_language":   b.getPrimaryLanguage(),
		"language_count":     len(b.languages),
		"language_diversity": b.calculateLanguageDiversity(),
	}
}

// GetCompleteAnalysis returns all metrics combined
func (b *AnalyzerDataBridge) GetCompleteAnalysis() map[string]interface{} {
	if b.IsEmpty() {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"repository":      b.GetRepositoryInfo(),
		"health":          b.GetHealthMetrics(),
		"contributors":    b.GetContributorMetrics(),
		"commits":         b.GetCommitMetrics(),
		"languages":       b.GetLanguageMetrics(),
		"summary":         b.GenerateSummary(),
		"recommendations": b.GenerateRecommendations(),
	}
}

// GetFileTree returns the repository file structure
func (b *AnalyzerDataBridge) GetFileTree() *FileNode {
	return b.fileTree
}
