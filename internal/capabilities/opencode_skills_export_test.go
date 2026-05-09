package capabilities

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteOpenCodeSkillsExport_DryRunNoWrites(t *testing.T) {
	target := t.TempDir()
	catalog, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	p, err := BuildOpenCodeExportProjection(catalog, []string{"sdr"})
	if err != nil {
		t.Fatalf("projection: %v", err)
	}
	res, err := ExecuteOpenCodeSkillsExport(p, target, true)
	if err != nil {
		t.Fatalf("dry-run export: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "bitsentry")); !os.IsNotExist(err) {
		t.Fatalf("dry run must not create bitsentry directory")
	}
	if res.Status != "preview" {
		t.Fatalf("expected preview status")
	}
}

func TestExecuteOpenCodeSkillsExport_RealWritesManagedAreaOnly(t *testing.T) {
	target := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	catalog, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	p, err := BuildOpenCodeExportProjection(catalog, []string{"sdr"})
	if err != nil {
		t.Fatalf("projection: %v", err)
	}
	res, err := ExecuteOpenCodeSkillsExport(p, target, false)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if res.Status != "exported" {
		t.Fatalf("expected exported status")
	}
	if _, err := os.Stat(filepath.Join(target, "bitsentry", "flows", "sdr.yaml")); err != nil {
		t.Fatalf("expected sdr flow exported: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "bitsentry", "flows", "sdd.yaml")); err == nil {
		t.Fatalf("did not expect sdd flow when only sdr selected")
	}
	if _, err := os.Stat(filepath.Join(target, "bitsentry", "skill-registry.md")); err != nil {
		t.Fatalf("expected generated skill-registry.md: %v", err)
	}
	if !strings.Contains(res.BackupPath, ".bitsentry-ai/backups/opencode-skills/") {
		t.Fatalf("expected backup path, got %s", res.BackupPath)
	}
}

func TestExecuteOpenCodeSkillsExport_PreventsTraversal(t *testing.T) {
	src := filepath.Join(t.TempDir(), "x.txt")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	p := OpenCodeExportProjection{
		SelectedIDs:    []string{"x"},
		IncludedSkills: []ProjectedSkill{{ID: "x/y", SourcePath: src, TargetPath: "../escape.txt", Status: "valid"}},
		GeneratedFiles: []ProjectedGeneratedFile{},
	}
	_, err := ExecuteOpenCodeSkillsExport(p, t.TempDir(), true)
	if err == nil {
		t.Fatalf("expected traversal protection error")
	}
}

func TestWriteOpenCodeSkillsExportReport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := WriteOpenCodeSkillsExportReport(SkillsExportResult{DryRun: true, Status: "preview", TargetRoot: "/tmp/x", SelectedIDs: []string{"sdr"}})
	if err != nil {
		t.Fatalf("write report: %v", err)
	}
	if !strings.Contains(path, ".bitsentry-ai/exports/opencode-skills/") {
		t.Fatalf("unexpected report path: %s", path)
	}
	if _, err := os.Stat(filepath.Join(home, ".bitsentry-ai", "exports", "opencode-skills", "latest.yaml")); err != nil {
		t.Fatalf("expected latest report: %v", err)
	}
}

func TestExecuteOpenCodeSkillsExport_DynamicFakeFlow(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	mustWrite(t, filepath.Join(root, "assets/flows/fake-flow.yaml"), "id: fake-flow\nname: Fake\nkind: flow\nselectable: true\ntop_level_flow: true\nfamily: fake\nskill_pack: fake-pack\norchestrator_skill: fake-pack/fake-orch\nstatus: available\ntriggers: [\"fake\"]\ncontracts: []\nrequires: {}\npersistence: {}\nstages: [{name: x, skill: fake-pack/fake-orch}]\nstage_graph: {}\nhandoffs: []\nfinal_artifacts: []\noutputs: []\n")
	mustWrite(t, filepath.Join(root, "assets/skills/fake-pack/fake-orch/SKILL.md"), validSkillContent("fake-orch"))
	mustWrite(t, filepath.Join(root, "assets/skills/_shared/a.md"), "# a\n")

	cat, err := DiscoverAssets(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	p, err := BuildOpenCodeExportProjection(cat, []string{"fake-flow"})
	if err != nil {
		t.Fatalf("projection: %v", err)
	}
	_, err = ExecuteOpenCodeSkillsExport(p, t.TempDir(), false)
	if err != nil {
		t.Fatalf("export dynamic: %v", err)
	}
}
