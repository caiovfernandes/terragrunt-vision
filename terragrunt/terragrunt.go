package terragrunt

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/caiovfernandes/terragrunt-runner/utils"
)

type File struct {
	Path      string
	Content   string
	RegionID  string
	ProjectID string
	StackID   string
}

type Workspace struct {
	Projects map[string]*Project
}

type Project struct {
	Name    string
	Regions map[string]*Region
}

type Region struct {
	Name   string
	Stacks map[string]*Stack
}

type Stack struct {
	Name  string
	Files []File
}

const baseFolder = "workspaces"
const minPathPartsLength = 4

func (h *Workspace) addFileToHierarchy(filePath string) {
	pathParts, baseIndex := extractPathParts(filePath, baseFolder)
	if baseIndex == -1 || len(pathParts) < baseIndex+minPathPartsLength {
		fmt.Printf("Skipping malformed path: %s\n", filePath)
		return
	}

	projectName, regionName, stackName := pathParts[baseIndex+1], pathParts[baseIndex+2], pathParts[baseIndex+3]
	_, _, stack := h.fetchOrCreateHierarchy(projectName, regionName, stackName)

	content, err := getFileContent(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	stack.Files = append(stack.Files, File{Path: filePath, Content: content, RegionID: regionName, ProjectID: projectName, StackID: stackName})
}

func extractPathParts(filePath, baseFolder string) ([]string, int) {
	pathParts := strings.Split(filePath, "/")
	baseIndex := -1
	for i, part := range pathParts {
		if part == baseFolder {
			baseIndex = i
			break
		}
	}
	return pathParts, baseIndex
}

func (h *Workspace) fetchOrCreateHierarchy(projectName, regionName, stackName string) (*Project, *Region, *Stack) {
	project := h.getOrCreateProject(projectName)
	region := project.getOrCreateRegion(regionName)
	stack := region.getOrCreateStack(stackName)
	return project, region, stack
}

func (h *Workspace) getOrCreateProject(name string) *Project {
	project, exists := h.Projects[name]
	if !exists {
		project = &Project{Regions: make(map[string]*Region)}
		h.Projects[name] = project
	}
	return project
}

func (p *Project) getOrCreateRegion(name string) *Region {
	region, exists := p.Regions[name]
	if !exists {
		region = &Region{Stacks: make(map[string]*Stack)}
		p.Regions[name] = region
	}
	return region
}

func (r *Region) getOrCreateStack(name string) *Stack {
	stack, exists := r.Stacks[name]
	if !exists {
		stack = &Stack{}
		r.Stacks[name] = stack
	}
	return stack
}

func (h *Workspace) PrintHierarchy() {
	for projectName, project := range h.Projects {
		fmt.Printf("Project: %s\n", projectName)
		for regionName, region := range project.Regions {
			fmt.Printf("  Region: %s\n", regionName)
			for stackName, stack := range region.Stacks {
				fmt.Printf("    Stack: %s\n", stackName)
				for _, file := range stack.Files {
					fmt.Printf("      File: %s\n", file.Path)
				}
			}
		}
	}
}

func GetWorkspace() (Workspace, error) {
	root := Workspace{Projects: make(map[string]*Project)}
	if len(os.Args) < 2 {
		return Workspace{}, errors.New("Usage: program <root-directory>")
	}
	rootDir := os.Args[1]
	fmt.Printf("rootDir: %s\n", rootDir)
	terragruntFiles, err := getTerragruntFiles(rootDir)
	if err != nil {
		return Workspace{}, err
	}
	for _, file := range terragruntFiles {
		root.addFileToHierarchy(file)
	}
	return root, nil
}

func getTerragruntFiles(rootDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "terragrunt.hcl") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func getFileContent(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(content), nil
}

func RunTerraformInit(rootDir string) (string, error) {
	// Remove the last item from the path
	parentDir := filepath.Dir(rootDir)
	accessKeyID, secretAccessKey, sessionToken, err := utils.GetAwsCredentials()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("terragrunt", "init", "-reconfigure")
	cmd.Env = append(os.Environ(),
		"AWS_ACCESS_KEY_ID="+accessKeyID,
		"AWS_SECRET_ACCESS_KEY="+secretAccessKey,
		"AWS_SESSION_TOKEN="+sessionToken,
	)
	cmd.Dir = parentDir

	outputBytes, err := cmd.CombinedOutput()
	outputFile := filepath.Join(cmd.Dir, "output")
	if err := ioutil.WriteFile(outputFile, outputBytes, 0644); err != nil {
		return string(outputBytes), fmt.Errorf("failed to save output to file: %v", err)
	}
	return string(outputBytes), err
}

func (h *Workspace) GetProjects() []string {
	projectMap := make(map[string]struct{})
	for project := range h.Projects {
		projectMap[project] = struct{}{}
	}

	projects := make([]string, 0, len(projectMap))
	for project := range projectMap {
		projects = append(projects, project)
	}
	return projects
}

func (h *Workspace) GetRegions() []string {
	regionMap := make(map[string]struct{})
	for _, project := range h.Projects {
		for region := range project.Regions {
			regionMap[region] = struct{}{}
		}
	}

	regions := make([]string, 0, len(regionMap))
	for region := range regionMap {
		regions = append(regions, region)
	}
	return regions
}

func (h *Workspace) GetStacks() []string {
	stackMap := make(map[string]struct{})
	for _, project := range h.Projects {
		for _, region := range project.Regions {
			for stack := range region.Stacks {
				stackMap[stack] = struct{}{}
			}
		}
	}

	stacks := make([]string, 0, len(stackMap))
	for stack := range stackMap {
		stacks = append(stacks, stack)
	}
	return stacks
}
