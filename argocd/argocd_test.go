package argocd_test

import (
	"os"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"roob.re/diffbot/argocd"
)

func Test_Applications(t *testing.T) {
	path := os.Getenv("TEST_MANUAL_PATH")
	if path == "" {
		t.Skip("Skipping manual test as TEST_MANUAL_PATH is not defined")
	}

	t.Log(argocd.Applications(path))
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
