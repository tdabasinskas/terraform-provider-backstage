package backstage

import (
	"os"
	"testing"
)

func TestAccResourceLocation(t *testing.T) {
	if os.Getenv("ACCTEST_SKIP_RESOURCE_TEST") != "" {
		t.Skip("Skipping as ACCTEST_SKIP_RESOURCE_LOCATION is set")
	}
}
