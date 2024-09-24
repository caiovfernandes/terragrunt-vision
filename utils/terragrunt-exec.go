package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	// Prepare the command
	cmd := exec.Command("terragrunt", "plan")

	// Set the working directory
	cmd.Dir = "/Users/caiofernandes/projects/prophecy/code/aws-prophecy-emite-infra/workspaces/"

	// Set the output to the standard output and error
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		log.Fatalf("command failed: %v", err)
	}

	fmt.Println("Command executed successfully.")
}
