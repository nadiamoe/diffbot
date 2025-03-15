package gitea_test

import (
	"os"
	"testing"

	"go.nadia.moe/diffbot/gitea"
)

func Test_Manual(t *testing.T) {
	t.Parallel()

	token := os.Getenv("TEST_GITEA_TOKEN")
	if token == "" {
		t.Skip("Skipping manual test as TEST_GITEA_TOKEN is not set")
	}

	err := gitea.PostOrUpdate(
		"https://git.tenshi.es",
		token,
		"k8s-manifests",
		"tenshi",
		"866",
		"Lorem ipsum dolor sit amet",
	)
	if err != nil {
		t.Fatalf("postorupdating: %v", err)
	}
}
