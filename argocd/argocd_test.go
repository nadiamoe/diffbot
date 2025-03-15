package argocd_test

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.nadia.moe/diffbot/argocd"
)

func Test_Applications(t *testing.T) {
	apps, err := argocd.Applications("testdata/applications/")
	if err != nil {
		t.Fatalf("parsing applications: %v", err)
	}

	for _, tc := range []struct {
		appName         string
		expectedSources []string
	}{
		{
			appName: "singlesource",
			expectedSources: []string{
				"nadia/workloads/singlesource",
			},
		},
		{
			appName:         "multisource",
			expectedSources: []string{},
		},
	} {
		appI := slices.IndexFunc(apps, func(app argocd.App) bool { return app.Name == tc.appName })
		if appI == -1 {
			t.Fatalf("%q app not found", tc.appName)
		}

		app := apps[appI]

		if app.SrcPath == "" && len(tc.expectedSources) != 0 {
			t.Fatalf("%q does not have the expected sources", tc.appName)
		}

		for _, expectedSource := range tc.expectedSources {
			if app.SrcPath != expectedSource {
				t.Fatalf("expected source %q not found in app", expectedSource)
			}
		}
	}
}

func Test_Changed(t *testing.T) {
	t.Parallel()

	apps := []argocd.App{
		{
			Name:    "test app",
			OwnPath: "dir/testapp.yaml",
			SrcPath: "dir/testapp",
		},
		{
			Name:    "test app 2",
			OwnPath: "dir/testapp2.yaml",
			SrcPath: "dir/testapp2",
		},
	}
	expectedApp := []string{"test app"}

	for _, tc := range []struct {
		name     string
		apps     []argocd.App
		files    []string
		expected []string
	}{
		{
			name:     "no match",
			apps:     apps,
			files:    []string{"random/file", "sub/dir/testapp.yaml", "sub/dir/testapp", "sub/dir/testapp/file", "dir2/testapp.yaml", "testapp.yaml", "testapp"},
			expected: nil,
		},
		{
			name:     "application spec changed",
			apps:     apps,
			files:    []string{"another/file", "dir/testapp.yaml"},
			expected: expectedApp,
		},
		{
			name:     "file inside source dir changed",
			apps:     apps,
			files:    []string{"another/file", "dir/testapp/something"},
			expected: expectedApp,
		},
		{
			name:     "several files changed for the same app",
			apps:     apps,
			files:    []string{"dir/testapp.yaml", "dir/testapp/something", "dir/testapp/something-else"},
			expected: expectedApp,
		},
		{
			name:     "two apps changed",
			apps:     apps,
			files:    []string{"another/file", "dir/testapp2.yaml", "dir/testapp/something"},
			expected: []string{"test app", "test app 2"},
		},
		{
			name: "does not confuse apps with common prefixes",
			apps: []argocd.App{
				{
					Name:    "foo",
					SrcPath: "foo",
					OwnPath: "foo.yaml",
				},
				{
					Name:    "foo2",
					SrcPath: "foo2",
					OwnPath: "foo2.yaml",
				},
			},
			files:    []string{"foo2.yaml"},
			expected: []string{"foo2"},
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := argocd.Changed(tc.apps, tc.files)
			slices.Sort(actual)

			if diff := cmp.Diff(actual, tc.expected); diff != "" {
				t.Fatalf("changed apps do not match expected:\n%s", diff)
			}
		})
	}
}
