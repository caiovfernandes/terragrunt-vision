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

// File represents a single Terragrunt file.
type File struct {
	Path    string
	Content string
}

// Stack represents a logical grouping of files.
type Stack struct {
	Files []File
}

// Region represents a geographical region (e.g., ap-southeast-2, us-west-2).
type Region struct {
	Stacks map[string]Stack
}

// Project represents a project (e.g., prophecy-dev, prophecy-prod).
type Project struct {
	Regions map[string]Region
}

// Root represents the overall structure holding all projects.
type Root struct {
	Projects map[string]Project
}

// addFileToHierarchy adds a file to the correct project, region, and stack in the hierarchy.
func (root *Root) addFileToHierarchy(filePath string) {
	// Define the base folder in the path after which we expect the project, region, and stack
	baseFolder := "workspaces"

	// Find the index of the "workspaces" directory in the path
	pathParts := strings.Split(filePath, "/")
	baseIdx := -1
	for i, part := range pathParts {
		if part == baseFolder {
			baseIdx = i
			break
		}
	}

	// Ensure the path has enough parts to extract project, region, and stack
	if baseIdx == -1 || len(pathParts) < baseIdx+4 {
		fmt.Printf("Skipping malformed path: %s\n", filePath)
		return
	}

	// Extract project, region, and stack from the path
	projectName := pathParts[baseIdx+1] // e.g., prophecy-dev
	regionName := pathParts[baseIdx+2]  // e.g., ap-southeast-2
	stackName := pathParts[baseIdx+3]   // e.g., acm, alb, etc.

	// Ensure the project exists, or create it
	project, projectExists := root.Projects[projectName]
	if !projectExists {
		project = Project{Regions: make(map[string]Region)}
		root.Projects[projectName] = project
	}

	// Ensure the region exists, or create it
	region, regionExists := project.Regions[regionName]
	if !regionExists {
		region = Region{Stacks: make(map[string]Stack)}
		project.Regions[regionName] = region
	}

	// Ensure the stack exists, or create it
	stack, stackExists := region.Stacks[stackName]
	if !stackExists {
		stack = Stack{}
	}

	content, err := getFileContent(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Add the file to the stack
	stack.Files = append(stack.Files, File{Path: filePath, Content: content})
	region.Stacks[stackName] = stack
	project.Regions[regionName] = region
	root.Projects[projectName] = project
}

// PrintHierarchy prints the hierarchy of projects, regions, stacks, and files.
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
	//  the root structure
	root := Root{Projects: make(map[string]Project)}

	if len(os.Args) < 2 {
		return nil, errors.New("Usage: program <root-directory>")
	}

	rootDir := os.Args[1]
	fmt.Printf("rootDir: %s\n", rootDir)

	// Get the terragrunt files recursively from the specified directory
	terragruntFiles, err := getTerragruntFiles(rootDir)
	if err != nil {
		return nil, err
	}

	// Add each file to the hierarchy
	for _, file := range terragruntFiles {
		root.addFileToHierarchy(file)
	}

	return root.Projects, nil

	// Print the hierarchy
	//root.PrintHierarchy() TODO: Improve debuggind adding this
}
func getTerragruntFiles(rootDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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
	cmd := exec.Command("terragrunt", "plan", "--terragrunt-forward-tf-stdout", "--terragrunt-non-interactive", "--terragrunt-log-level error", "--terragrunt-log-disable")
	cmd.Dir = "/Users/caiofernandes/projects/prophecy/code/aws-prophecy-emite-infra/workspaces/prophecy-dev/ap-southeast-2/alb/emite-alb/"
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return string(outputBytes), err
	}
	//output := string(outputBytes)
	output := cmd.String()

	// Save output to a file
	//outputFile := filepath.Join(cmd.Dir, "output")
	//if err := ioutil.WriteFile(outputFile, []byte(output), 0644); err != nil {
	//	return output, fmt.Errorf("failed to save output to file: %v", err)
	//}

	return output, nil
}
