package terragrunt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceHierarchy(t *testing.T) {
	workspace := Workspace{Projects: make(map[string]*Project)}

	// Test adding a file to hierarchy
	testPath := "workspaces/test-project/us-east-1/vpc/terragrunt.hcl"
	workspace.addFileToHierarchy(testPath)

	// Verify project was created
	if len(workspace.Projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(workspace.Projects))
	}

	project, exists := workspace.Projects["test-project"]
	if !exists {
		t.Fatal("Expected test-project to exist")
	}

	// Verify region was created
	if len(project.Regions) != 1 {
		t.Errorf("Expected 1 region, got %d", len(project.Regions))
	}

	region, exists := project.Regions["us-east-1"]
	if !exists {
		t.Fatal("Expected us-east-1 region to exist")
	}

	// Verify stack was created
	if len(region.Stacks) != 1 {
		t.Errorf("Expected 1 stack, got %d", len(region.Stacks))
	}

	_, exists = region.Stacks["vpc"]
	if !exists {
		t.Fatal("Expected vpc stack to exist")
	}
}

func TestGetProjects(t *testing.T) {
	workspace := Workspace{Projects: make(map[string]*Project)}
	workspace.Projects["project1"] = &Project{Name: "project1", Regions: make(map[string]*Region)}
	workspace.Projects["project2"] = &Project{Name: "project2", Regions: make(map[string]*Region)}

	projects := workspace.GetProjects()
	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}
}

func TestGetRegions(t *testing.T) {
	workspace := Workspace{Projects: make(map[string]*Project)}
	project := &Project{Name: "test", Regions: make(map[string]*Region)}
	project.Regions["us-east-1"] = &Region{Name: "us-east-1", Stacks: make(map[string]*Stack)}
	project.Regions["us-west-2"] = &Region{Name: "us-west-2", Stacks: make(map[string]*Stack)}
	workspace.Projects["test"] = project

	regions := workspace.GetRegions()
	if len(regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(regions))
	}
}

func TestGetStacks(t *testing.T) {
	workspace := Workspace{Projects: make(map[string]*Project)}
	project := &Project{Name: "test", Regions: make(map[string]*Region)}
	region := &Region{Name: "us-east-1", Stacks: make(map[string]*Stack)}
	region.Stacks["vpc"] = &Stack{Name: "vpc"}
	region.Stacks["ec2"] = &Stack{Name: "ec2"}
	project.Regions["us-east-1"] = region
	workspace.Projects["test"] = project

	stacks := workspace.GetStacks()
	if len(stacks) != 2 {
		t.Errorf("Expected 2 stacks, got %d", len(stacks))
	}
}

func TestExtractPathParts(t *testing.T) {
	tests := []struct {
		path       string
		baseFolder string
		wantIndex  int
	}{
		{"workspaces/project/region/stack/terragrunt.hcl", "workspaces", 0},
		{"/home/user/workspaces/project/region/stack/terragrunt.hcl", "workspaces", 3},
		{"no/match/here/terragrunt.hcl", "workspaces", -1},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			_, index := extractPathParts(tt.path, tt.baseFolder)
			if index != tt.wantIndex {
				t.Errorf("extractPathParts(%q, %q) index = %d, want %d", tt.path, tt.baseFolder, index, tt.wantIndex)
			}
		})
	}
}

func TestGetTerragruntFiles(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "terragrunt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{
		filepath.Join(tmpDir, "workspaces", "proj1", "us-east-1", "vpc", "terragrunt.hcl"),
		filepath.Join(tmpDir, "workspaces", "proj1", "us-east-1", "ec2", "terragrunt.hcl"),
		filepath.Join(tmpDir, "workspaces", "proj2", "us-west-2", "rds", "terragrunt.hcl"),
	}

	for _, file := range testFiles {
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(file, []byte("# test content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Test getTerragruntFiles
	files, err := getTerragruntFiles(tmpDir)
	if err != nil {
		t.Fatalf("getTerragruntFiles failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 terragrunt files, got %d", len(files))
	}
}
