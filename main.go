package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"roob.re/diffbot/argocd"
	"roob.re/diffbot/gitea"
)

func main() {
	if len(os.Args) >= 2 {
		_ = os.Chdir(os.Args[1])
	}

	changedFilesEnv := os.Getenv("CI_PIPELINE_FILES")
	log.Printf("Using list of changed files supplied by CI: %s", changedFilesEnv)

	var changedFiles []string
	err := json.Unmarshal([]byte(changedFilesEnv), &changedFiles)
	if err != nil {
		log.Printf("Unmarshalling json from changed files env var: %v", err)
	}

	if len(changedFiles) == 0 {
		log.Printf("No changed files found")
		return
	}

	applications, err := argocd.Applications(".")
	if err != nil {
		log.Fatalf("listing applications: %v", err)
	}

	if len(applications) == 0 {
		log.Print("No argocd applications found")
		return
	}

	changedApps := argocd.Changed(applications, changedFiles)
	if len(changedApps) == 0 {
		log.Print("No changes detected")
		return
	}

	log.Printf("Detected changes in: %s", strings.Join(changedApps, ", "))

	commentBuf := &bytes.Buffer{}

	for _, changedApp := range changedApps {
		log.Printf("Running argocd diff for %s", changedApp)
		diff, err := argocd.Diff(changedApp)
		if err != nil {
			log.Printf("getting argo diff for %q: %v", changedApp, err)
			continue
		}

		fmt.Fprintf(commentBuf, "## Changes in `%s`\n\n```diff\n%s\n```\n\n", changedApp, strings.TrimSpace(diff))
	}

	comment := fmt.Sprintf("Diff generated on %v (%s)\n\n%s", time.Now(), os.Getenv("CI_COMMIT_SHA"), commentBuf.String())
	err = gitea.PostOrUpdate(
		os.Getenv("CI_FORGE_URL"),
		os.Getenv("GITEA_TOKEN"),
		os.Getenv("CI_REPO_OWNER"),
		os.Getenv("CI_REPO_NAME"),
		os.Getenv("CI_COMMIT_PULL_REQUEST"),
		comment,
	)
	if err != nil {
		log.Printf("Error posting comment: %v", err)
		return
	}
}

func gitChanged(base string) ([]string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command("git", "diff", "--name-only", base)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("running git diff: %w\n%s", err, stderr.String())
	}

	return strings.Split(stdout.String(), "\n"), nil
}
