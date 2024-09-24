package terragrunt

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type File struct {
	Path    string
	Content string
}

type Stack struct {
	Files []File
}

type Region struct {
	Stacks map[string]Stack
}

type Project struct {
	Regions map[string]Region
}

type Root struct {
	Projects map[string]Project
}

func (root *Root) addFileToHierarchy(filePath string) {
	baseFolder := "workspaces"

	pathParts := strings.Split(filePath, "/")
	baseIdx := -1
	for i, part := range pathParts {
		if part == baseFolder {
			baseIdx = i
			break
		}
	}

	if baseIdx == -1 || len(pathParts) < baseIdx+4 {
		fmt.Printf("Skipping malformed path: %s\n", filePath)
		return
	}

	projectName := pathParts[baseIdx+1] // e.g., prophecy-dev
	regionName := pathParts[baseIdx+2]  // e.g., ap-southeast-2
	stackName := pathParts[baseIdx+3]   // e.g., acm, alb, etc.

	project, projectExists := root.Projects[projectName]
	if !projectExists {
		project = Project{Regions: make(map[string]Region)}
		root.Projects[projectName] = project
	}

	region, regionExists := project.Regions[regionName]
	if !regionExists {
		region = Region{Stacks: make(map[string]Stack)}
		project.Regions[regionName] = region
	}

	stack, stackExists := region.Stacks[stackName]
	if !stackExists {
		stack = Stack{}
	}

	content, err := getFileContent(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	stack.Files = append(stack.Files, File{Path: filePath, Content: content})
	region.Stacks[stackName] = stack
	project.Regions[regionName] = region
	root.Projects[projectName] = project
}

func (root *Root) PrintHierarchy() {
	for projectName, project := range root.Projects {
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

func GetProjects() (map[string]Project, error) {
	root := Root{Projects: make(map[string]Project)}

	if len(os.Args) < 2 {
		return nil, errors.New("Usage: program <root-directory>")
	}

	rootDir := os.Args[1]
	fmt.Printf("rootDir: %s\n", rootDir)

	terragruntFiles, err := getTerragruntFiles(rootDir)
	if err != nil {
		return nil, err
	}

	for _, file := range terragruntFiles {
		root.addFileToHierarchy(file)
	}

	return root.Projects, nil

	// Print the hierarchy
	// root.PrintHierarchy() TODO: Improve debuggind adding this
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

func getFileContent(file_path string) (string, error) {
	content, err := ioutil.ReadFile(file_path)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(content), nil
}

func RunTerraformInit(rootDir string) (string, error) {
	cmd := exec.Command("terragrunt", "plan", "-reconfigure", "-terragrunt-forward-tf-stdout", "-terragrunt-non-interactive")
	cmd.Dir = "/Users/caiofernandes/projects/prophecy/code/aws-prophecy-emite-infra/workspaces/prophecy-dev/ap-southeast-2/alb/emite-alb/"
	outputBytes, err := cmd.CombinedOutput()
	// Save output to a file
	outputFile := filepath.Join(cmd.Dir, "output")
	if err := ioutil.WriteFile(outputFile, outputBytes, 0644); err != nil {
		return string(outputBytes), fmt.Errorf("failed to save output to file: %v", err)
	}

	return string(outputBytes), err
}
