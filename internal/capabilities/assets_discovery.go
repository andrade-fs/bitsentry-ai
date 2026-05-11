package capabilities

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var requiredSkillHeadings = []string{
	"## Purpose",
	"## Use When",
	"## Inputs",
	"## Workflow",
	"## Outputs",
	"## Boundaries",
	"## Persistence Actions",
	"## Result Envelope",
	"## Handoffs",
	"## Quality Checklist",
}

type AssetCatalog struct {
	Flows         []DiscoveredFlow
	SkillPacks    []DiscoveredSkillPack
	Skills        []DiscoveredSkill
	Shared        []DiscoveredSharedContract
	Orchestrators []DiscoveredOrchestrator
	Intents       []DiscoveredIntent
	Roles         []DiscoveredRole
}

type DiscoveredIntent struct {
	ID                     string   `yaml:"id"`
	Description            string   `yaml:"description"`
	DefaultDecision        string   `yaml:"default_decision"`
	DefaultFlow            string   `yaml:"default_flow"`
	AlternativeFlow        string   `yaml:"alternative_flow"`
	ComplexityThreshold    string   `yaml:"complexity_threshold"`
	PreFlowRoles           []string `yaml:"pre_flow_roles"`
	PreFlowSkills          []string `yaml:"pre_flow_skills"`
	ExpectedContextOutputs []string `yaml:"expected_context_outputs"`
	DirectAnswerAllowed    bool     `yaml:"direct_answer_allowed"`
	RequiresConfirmation   bool     `yaml:"requires_confirmation"`
	RequiresBoundedDiscovery bool   `yaml:"requires_bounded_discovery"`
	ForbiddenActions       []string `yaml:"forbidden_actions"`
	SourcePath             string
}

type roleFrontmatter struct {
	ID          string            `yaml:"id"`
	Category    string            `yaml:"category"`
	Kind        string            `yaml:"kind"`
	UsableIn    []string          `yaml:"usable_in"`
	Permissions map[string]string `yaml:"permissions"`
}

type DiscoveredRole struct {
	ID          string
	Category    string
	Kind        string
	UsableIn    []string
	Permissions map[string]string
	Title       string
	SourcePath  string
	NonEmpty    bool
}

type DiscoveredFlow struct {
	ID                string            `yaml:"id"`
	Name              string            `yaml:"name"`
	Kind              string            `yaml:"kind"`
	Selectable        bool              `yaml:"selectable"`
	TopLevelFlow      bool              `yaml:"top_level_flow"`
	Family            string            `yaml:"family"`
	SkillPack         string            `yaml:"skill_pack"`
	OrchestratorSkill string            `yaml:"orchestrator_skill"`
	Status            string            `yaml:"status"`
	Triggers          []string          `yaml:"triggers"`
	Contracts         []string          `yaml:"contracts"`
	Requires          map[string]any    `yaml:"requires"`
	Persistence       map[string]string `yaml:"persistence"`
	Stages            []map[string]any  `yaml:"stages"`
	StageGraph        map[string]any    `yaml:"stage_graph"`
	Handoffs          []map[string]any  `yaml:"handoffs"`
	FinalArtifacts    []any             `yaml:"final_artifacts"`
	Outputs           []any             `yaml:"outputs"`
	SourcePath        string
}

type DiscoveredSkillPack struct {
	ID             string
	SourcePath     string
	SkillCount     int
	SkillFiles     []string
	Shared         bool
	RelatedFlowIDs []string
}

type DiscoveredSkill struct {
	ID                      string
	PackID                  string
	RelativePath            string
	SourcePath              string
	Title                   string
	FrontmatterName         string
	Description             string
	RequiredHeadingsPresent []string
	RequiredHeadingsMissing []string
	Status                  string
}

type DiscoveredSharedContract struct {
	ID       string
	Path     string
	NonEmpty bool
}

type DiscoveredOrchestrator struct {
	ID       string
	Path     string
	Title    string
	NonEmpty bool
}

type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description any    `yaml:"description"`
}

func DiscoverAssets(root string) (AssetCatalog, error) {
	assetsRoot, err := resolveAssetsRoot(root)
	if err != nil {
		return AssetCatalog{}, err
	}
	flows, err := discoverFlows(assetsRoot)
	if err != nil {
		return AssetCatalog{}, err
	}
	shared, err := discoverSharedContracts(assetsRoot)
	if err != nil {
		return AssetCatalog{}, err
	}
	skillPacks, skills, err := discoverSkillPacksAndSkills(assetsRoot, flows)
	if err != nil {
		return AssetCatalog{}, err
	}
	orchestrators, err := discoverOrchestrators(assetsRoot)
	if err != nil {
		return AssetCatalog{}, err
	}
	intents, err := discoverIntents(assetsRoot)
	if err != nil {
		return AssetCatalog{}, err
	}
	roles, err := discoverRoles(assetsRoot)
	if err != nil {
		return AssetCatalog{}, err
	}
	return AssetCatalog{
		Flows:         flows,
		SkillPacks:    skillPacks,
		Skills:        skills,
		Shared:        shared,
		Orchestrators: orchestrators,
		Intents:       intents,
		Roles:         roles,
	}, nil
}

func resolveAssetsRoot(root string) (string, error) {
	trimmed := strings.TrimSpace(root)
	if trimmed == "" {
		trimmed = "."
	}
	absRoot, err := filepath.Abs(trimmed)
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(absRoot)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("discovery root must be a directory: %s", absRoot)
	}
	if filepath.Base(absRoot) == "assets" {
		return absRoot, nil
	}
	assetsPath := filepath.Join(absRoot, "assets")
	if st, err := os.Stat(assetsPath); err == nil && st.IsDir() {
		return assetsPath, nil
	}
	return "", fmt.Errorf("assets directory not found from root: %s", absRoot)
}

func discoverFlows(assetsRoot string) ([]DiscoveredFlow, error) {
	entries, err := filepath.Glob(filepath.Join(assetsRoot, "flows", "*.yaml"))
	if err != nil {
		return nil, err
	}
	flows := make([]DiscoveredFlow, 0, len(entries))
	for _, p := range entries {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var f DiscoveredFlow
		if err := yaml.Unmarshal(raw, &f); err != nil {
			return nil, err
		}
		f.SourcePath = p
		flows = append(flows, f)
	}
	sort.Slice(flows, func(i, j int) bool { return flows[i].ID < flows[j].ID })
	return flows, nil
}

func discoverSharedContracts(assetsRoot string) ([]DiscoveredSharedContract, error) {
	entries, err := filepath.Glob(filepath.Join(assetsRoot, "skills", "_shared", "*.md"))
	if err != nil {
		return nil, err
	}
	out := make([]DiscoveredSharedContract, 0, len(entries))
	for _, p := range entries {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		out = append(out, DiscoveredSharedContract{
			ID:       strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)),
			Path:     p,
			NonEmpty: strings.TrimSpace(string(raw)) != "",
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func discoverSkillPacksAndSkills(assetsRoot string, flows []DiscoveredFlow) ([]DiscoveredSkillPack, []DiscoveredSkill, error) {
	root := filepath.Join(assetsRoot, "skills")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil, err
	}
	related := map[string]map[string]bool{}
	for _, f := range flows {
		if strings.TrimSpace(f.SkillPack) == "" || strings.TrimSpace(f.ID) == "" {
			continue
		}
		if _, ok := related[f.SkillPack]; !ok {
			related[f.SkillPack] = map[string]bool{}
		}
		related[f.SkillPack][f.ID] = true
	}
	packs := make([]DiscoveredSkillPack, 0, len(entries))
	allSkills := make([]DiscoveredSkill, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		packID := e.Name()
		packPath := filepath.Join(root, packID)
		isShared := packID == "_shared"
		skillFiles, skills, err := discoverSkillsInPack(packID, packPath, assetsRoot)
		if err != nil {
			return nil, nil, err
		}
		allSkills = append(allSkills, skills...)
		relatedFlowIDs := make([]string, 0)
		if r, ok := related[packID]; ok {
			for id := range r {
				relatedFlowIDs = append(relatedFlowIDs, id)
			}
			sort.Strings(relatedFlowIDs)
		}
		packs = append(packs, DiscoveredSkillPack{
			ID:             packID,
			SourcePath:     packPath,
			SkillCount:     len(skillFiles),
			SkillFiles:     skillFiles,
			Shared:         isShared,
			RelatedFlowIDs: relatedFlowIDs,
		})
	}
	sort.Slice(packs, func(i, j int) bool { return packs[i].ID < packs[j].ID })
	sort.Slice(allSkills, func(i, j int) bool { return allSkills[i].ID < allSkills[j].ID })
	return packs, allSkills, nil
}

func discoverSkillsInPack(packID, packPath, assetsRoot string) ([]string, []DiscoveredSkill, error) {
	skillFiles := []string{}
	skills := []DiscoveredSkill{}
	err := filepath.WalkDir(packPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(assetsRoot, path)
		if err != nil {
			return err
		}
		text := string(raw)
		present := []string{}
		missing := []string{}
		for _, h := range requiredSkillHeadings {
			if hasRequiredHeading(text, h) {
				present = append(present, h)
			} else {
				missing = append(missing, h)
			}
		}
		name, desc := parseSkillFrontmatter(raw)
		title := parseSkillTitle(text)
		relSlash := filepath.ToSlash(rel)
		skillID := skillIDFromPath(packID, relSlash)
		status := "valid"
		if len(missing) > 0 {
			status = "invalid"
		}
		skillFiles = append(skillFiles, path)
		skills = append(skills, DiscoveredSkill{
			ID:                      skillID,
			PackID:                  packID,
			RelativePath:            relSlash,
			SourcePath:              path,
			Title:                   title,
			FrontmatterName:         name,
			Description:             desc,
			RequiredHeadingsPresent: present,
			RequiredHeadingsMissing: missing,
			Status:                  status,
		})
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(skillFiles)
	return skillFiles, skills, nil
}

func discoverOrchestrators(assetsRoot string) ([]DiscoveredOrchestrator, error) {
	dir := filepath.Join(assetsRoot, "orchestrators")
	st, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []DiscoveredOrchestrator{}, nil
		}
		return nil, err
	}
	if !st.IsDir() {
		return []DiscoveredOrchestrator{}, nil
	}
	entries, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	out := make([]DiscoveredOrchestrator, 0, len(entries))
	for _, p := range entries {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		text := string(raw)
		out = append(out, DiscoveredOrchestrator{
			ID:       strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)),
			Path:     p,
			Title:    parseMarkdownTitle(text),
			NonEmpty: strings.TrimSpace(text) != "",
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func discoverIntents(assetsRoot string) ([]DiscoveredIntent, error) {
	entries, err := filepath.Glob(filepath.Join(assetsRoot, "intents", "*.yaml"))
	if err != nil {
		return nil, err
	}
	out := make([]DiscoveredIntent, 0, len(entries))
	for _, p := range entries {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var in DiscoveredIntent
		if err := yaml.Unmarshal(raw, &in); err != nil {
			return nil, err
		}
		in.SourcePath = p
		out = append(out, in)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func discoverRoles(assetsRoot string) ([]DiscoveredRole, error) {
	entries, err := filepath.Glob(filepath.Join(assetsRoot, "roles", "*.md"))
	if err != nil {
		return nil, err
	}
	out := make([]DiscoveredRole, 0, len(entries))
	for _, p := range entries {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		fm := parseRoleFrontmatter(raw)
		text := string(raw)
		out = append(out, DiscoveredRole{
			ID:          strings.TrimSpace(fm.ID),
			Category:    strings.TrimSpace(fm.Category),
			Kind:        strings.TrimSpace(fm.Kind),
			UsableIn:    append([]string{}, fm.UsableIn...),
			Permissions: fm.Permissions,
			Title:       parseMarkdownTitle(text),
			SourcePath:  p,
			NonEmpty:    strings.TrimSpace(text) != "",
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func parseSkillTitle(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# Skill:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# Skill:"))
		}
	}
	return ""
}

func hasExactHeadingLine(text string, heading string) bool {
	required := strings.TrimSpace(heading)
	if required == "" {
		return false
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == required {
			return true
		}
	}
	return false
}

func hasRequiredHeading(text string, heading string) bool {
	return hasExactHeadingLine(text, heading)
}

func parseMarkdownTitle(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}
	return ""
}

func parseSkillFrontmatter(raw []byte) (name string, description string) {
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") {
		return "", ""
	}
	parts := strings.SplitN(text, "\n---\n", 2)
	if len(parts) != 2 {
		return "", ""
	}
	block := strings.TrimPrefix(parts[0], "---\n")
	var fm skillFrontmatter
	if err := yaml.Unmarshal([]byte(block), &fm); err != nil {
		return "", ""
	}
	return strings.TrimSpace(fm.Name), strings.TrimSpace(asString(fm.Description))
}

func parseRoleFrontmatter(raw []byte) roleFrontmatter {
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") {
		return roleFrontmatter{}
	}
	parts := strings.SplitN(text, "\n---\n", 2)
	if len(parts) != 2 {
		return roleFrontmatter{}
	}
	block := strings.TrimPrefix(parts[0], "---\n")
	var fm roleFrontmatter
	if err := yaml.Unmarshal([]byte(block), &fm); err != nil {
		return roleFrontmatter{}
	}
	if fm.Permissions == nil {
		fm.Permissions = map[string]string{}
	}
	return fm
}

func asString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

func skillIDFromPath(packID, relSlash string) string {
	prefix := "skills/"
	remainder := relSlash
	if strings.HasPrefix(relSlash, prefix) {
		remainder = strings.TrimPrefix(relSlash, prefix)
	}
	remainder = strings.TrimSuffix(remainder, "/SKILL.md")
	segs := strings.Split(remainder, "/")
	if len(segs) >= 2 {
		return fmt.Sprintf("%s/%s", packID, segs[len(segs)-1])
	}
	return fmt.Sprintf("%s/%s", packID, strings.TrimSuffix(filepath.Base(relSlash), ".md"))
}
