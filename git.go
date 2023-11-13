package main

import (
	"os/exec"
	"strings"
)

// Commit represents a Git commit
type Commit struct {
	Hash    string
	Author  string
	Date    string
	Message string
	Patch   string
}

func gitFetchFileHistory(repoPath, filePath string) ([]Commit, error){
	var commits []Commit
	var currentCommit Commit
	var isPatch bool

	cmd := exec.Command("git", "-C", repoPath, "log", "-p", filePath)
	logOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(logOutput), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "commit ") {
			if currentCommit.Hash != "" {
				commits = append(commits, currentCommit)
			}
			currentCommit = Commit{Hash: strings.TrimPrefix(line, "commit ")}
			isPatch = false
		} else if strings.HasPrefix(line, "Author: ") {
			currentCommit.Author = strings.TrimPrefix(line, "Author: ")
		} else if strings.HasPrefix(line, "Date:   ") {
			currentCommit.Date = strings.TrimPrefix(line, "Date:   ")
		} else if strings.HasPrefix(line, "diff --git ") {
			isPatch = true
			currentCommit.Patch += line + "\n"
		} else if isPatch {
			currentCommit.Patch += line + "\n"
		} else {
			currentCommit.Message += line + "\n"
		}
	}

	if currentCommit.Hash != "" {
		commits = append(commits, currentCommit)
	}

	return commits, nil
}
