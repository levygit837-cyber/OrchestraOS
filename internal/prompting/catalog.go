package prompting

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed catalog/manifest.json catalog/fragments
var catalogFS embed.FS

type fragmentManifestEntry struct {
	ID               string           `json:"id"`
	Version          string           `json:"version"`
	Category         FragmentCategory `json:"category"`
	Kind             FragmentKind     `json:"kind"`
	Title            string           `json:"title"`
	Priority         int              `json:"priority"`
	ExclusiveGroup   string           `json:"exclusive_group"`
	BodyPath         string           `json:"body_path"`
	BodyHash         string           `json:"body_hash"`
	AppliesWhen      AppliesWhen      `json:"applies_when"`
	Requires         []string         `json:"requires"`
	ConflictsWith    []string         `json:"conflicts_with"`
	Allows           []string         `json:"allows"`
	Denies           []string         `json:"denies"`
	ApprovalRequired []string         `json:"approval_required"`
	AutonomyLevel    int              `json:"autonomy_level"`
}

type Catalog struct {
	fragments []Fragment
}

func LoadCatalog() (*Catalog, error) {
	return loadCatalog(catalogFS)
}

func loadCatalog(source fs.FS) (*Catalog, error) {
	raw, err := fs.ReadFile(source, "catalog/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("read prompt fragment manifest: %w", err)
	}

	var entries []fragmentManifestEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("decode prompt fragment manifest: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("prompt fragment manifest is empty")
	}

	fragments := make([]Fragment, 0, len(entries))
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if entry.ID == "" || entry.Version == "" || entry.Category == "" || entry.Kind == "" || entry.Title == "" || entry.BodyPath == "" || entry.BodyHash == "" {
			return nil, fmt.Errorf("fragment manifest entry %q is missing required metadata", entry.ID)
		}
		key := entry.ID + "@" + entry.Version
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("duplicate prompt fragment %s", key)
		}
		seen[key] = struct{}{}

		bodyBytes, err := fs.ReadFile(source, "catalog/"+entry.BodyPath)
		if err != nil {
			return nil, fmt.Errorf("read prompt fragment %s: %w", entry.ID, err)
		}
		body := string(bodyBytes)
		hash := HashText(body)
		if hash != entry.BodyHash {
			return nil, fmt.Errorf("prompt fragment %s body hash mismatch: manifest %s, actual %s", entry.ID, entry.BodyHash, hash)
		}
		metadataHash, err := metadataHashForEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("hash prompt fragment metadata %s: %w", entry.ID, err)
		}
		fragments = append(fragments, Fragment{
			ID:               entry.ID,
			Version:          entry.Version,
			Category:         entry.Category,
			Kind:             entry.Kind,
			Title:            entry.Title,
			Priority:         entry.Priority,
			ExclusiveGroup:   entry.ExclusiveGroup,
			BodyPath:         entry.BodyPath,
			BodyHash:         entry.BodyHash,
			MetadataHash:     metadataHash,
			Body:             body,
			AppliesWhen:      entry.AppliesWhen,
			Requires:         entry.Requires,
			ConflictsWith:    entry.ConflictsWith,
			Allows:           entry.Allows,
			Denies:           entry.Denies,
			ApprovalRequired: entry.ApprovalRequired,
			AutonomyLevel:    entry.AutonomyLevel,
		})
	}

	return &Catalog{fragments: fragments}, nil
}

func (c *Catalog) Fragments() []Fragment {
	out := make([]Fragment, len(c.fragments))
	copy(out, c.fragments)
	return out
}

func HashText(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func metadataHashForEntry(entry fragmentManifestEntry) (string, error) {
	payload := struct {
		Kind             FragmentKind     `json:"kind"`
		Category         FragmentCategory `json:"category"`
		Priority         int              `json:"priority"`
		ExclusiveGroup   string           `json:"exclusive_group"`
		AppliesWhen      AppliesWhen      `json:"applies_when"`
		Requires         []string         `json:"requires"`
		ConflictsWith    []string         `json:"conflicts_with"`
		Allows           []string         `json:"allows"`
		Denies           []string         `json:"denies"`
		ApprovalRequired []string         `json:"approval_required"`
		AutonomyLevel    int              `json:"autonomy_level"`
		BodyHash         string           `json:"body_hash"`
	}{
		Kind:             entry.Kind,
		Category:         entry.Category,
		Priority:         entry.Priority,
		ExclusiveGroup:   entry.ExclusiveGroup,
		AppliesWhen:      normalizeAppliesWhen(entry.AppliesWhen),
		Requires:         sortedCopy(entry.Requires),
		ConflictsWith:    sortedCopy(entry.ConflictsWith),
		Allows:           sortedCopy(entry.Allows),
		Denies:           sortedCopy(entry.Denies),
		ApprovalRequired: sortedCopy(entry.ApprovalRequired),
		AutonomyLevel:    entry.AutonomyLevel,
		BodyHash:         entry.BodyHash,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return HashText(string(raw)), nil
}

func normalizeAppliesWhen(applies AppliesWhen) AppliesWhen {
	return AppliesWhen{AgentProfiles: sortedCopy(applies.AgentProfiles)}
}

func sortedCopy(values []string) []string {
	out := append([]string{}, values...)
	sort.Strings(out)
	if out == nil {
		return []string{}
	}
	return out
}

func normalizeProfile(profile string) string {
	profile = strings.TrimSpace(strings.ToLower(profile))
	if profile == "" {
		return "code_worker"
	}
	switch profile {
	case "default", "codex", "general_engineering":
		return "code_worker"
	}
	return profile
}
