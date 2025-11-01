package argocd

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
			sourceIncludesFile := func(appSource string) bool {
				return strings.HasPrefix(file, strings.TrimSuffix(appSource, "/")+"/")
			}

			if file == app.OwnPath || slices.ContainsFunc(app.SrcPaths, sourceIncludesFile) {
				log.Printf("file %q belongs to %q", file, app.Name)
				changed[app.Name] = struct{}{}
				continue appLoop
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

// Applications walks the `root` directory and loads all paths referenced by  argoproj.io/v1alpha1/Application manifests
// from yaml files found in that directory, recursively.
// `filterRepo` can be used to leave out paths based on the `repoURL` field of each Application's `source`. It will be
// called for each `source` using that `repoURL` as an argument, and the source will be included only if `filterRepo`
// returns true. A nil `filterRepo` causes everything to be included.
func Applications(root string, filterRepo func(string) bool) ([]App, error) {
	var apps []App
	fErr := filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		type sourceSpec struct {
			Path    string
			RepoURL string `yaml:"repoURL"`
		}
		var app struct {
			APIVersion string `yaml:"apiVersion"`
			Kind       string
			Metadata   struct {
				Name string
			}
			Spec struct {
				Source  sourceSpec
				Sources []sourceSpec
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
		for _, src := range append(app.Spec.Sources, app.Spec.Source) {
			if filterRepo != nil && !filterRepo(src.RepoURL) {
				continue
			}

			if src.Path == "" {
				continue
			}

			paths = append(paths, src.Path)
		}

		if len(paths) == 0 {
			// App has no sources, or no source matched filterRepo. Skip it.
			return nil
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
