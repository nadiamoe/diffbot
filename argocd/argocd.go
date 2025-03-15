package argocd

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const maxLines = 200

func Diff(appName string) (string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	realpath, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("getting absolute path of CWD: %w", err)
	}

	cmd := exec.Command("argocd", "app", "diff", "--server-side-generate", "--local", realpath, appName)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		exitErr, isExit := err.(*exec.ExitError)

		// Argocd exits with 1 when there is a diff. That is not an error.
		if !(isExit && exitErr.ExitCode() == 1) {
			return "", fmt.Errorf("running argocd diff: %w\n%s", err, stderr.String())
		}
	}

	stdoutString := stdout.String()
	if strings.Count(stdoutString, "\n") <= maxLines {
		return stdoutString, nil
	}

	output := &strings.Builder{}
	_, _ = fmt.Fprintf(output, "# %d more lines truncated from output\n\n", len(stdoutString)-maxLines)
	_, _ = output.WriteString(stdoutString)

	return output.String(), nil
}

func Changed(applications []App, changedFiles []string) []string {
	changed := map[string]struct{}{}

appLoop:
	for _, app := range applications {
		for _, file := range changedFiles {
			// A file belongs to an application if it is the path of the application yaml, or if it is in a subpath of
			// the application source path.
			// We add the trailing slash to prevent too ensure we match files inside the srcPath *directory*, not having
			// srcPath as a prefix: e.g. path "dir/foo2.yaml" being associated with a different app that has a source path of "dir/foo"
			for _, srcPath := range app.SrcPaths {
				if file == app.OwnPath || strings.HasPrefix(file, strings.TrimSuffix(srcPath, "/")+"/") {
					log.Printf("file %q belongs to %q", file, app.Name)
					changed[app.Name] = struct{}{}
					continue appLoop
				}
			}
		}
	}

	if len(changed) == 0 {
		return nil
	}

	changedList := make([]string, 0, len(changed))
	for app := range changed {
		changedList = append(changedList, app)
	}

	return changedList
}

type App struct {
	Name     string
	SrcPaths []string
	OwnPath  string
}

func Applications(root string) ([]App, error) {
	var apps []App
	fErr := filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		var app struct {
			APIVersion string `yaml:"apiVersion"`
			Kind       string
			Metadata   struct {
				Name string
			}
			Spec struct {
				Source struct {
					Path string
				}
			}
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		err = yaml.NewDecoder(file).Decode(&app)
		if err != nil {
			// Not the yaml we're looking for.
			return nil
		}

		if !strings.HasPrefix(app.APIVersion, "argoproj.io/v1") {
			return nil
		}

		if !strings.EqualFold(app.Kind, "application") {
			return nil
		}

		var paths []string
		if app.Spec.Source.Path != "" {
			paths = append(paths, app.Spec.Source.Path)
		}

		apps = append(apps, App{
			Name:     app.Metadata.Name,
			SrcPaths: paths,
			OwnPath:  path,
		})

		return nil
	})

	return apps, fErr
}
