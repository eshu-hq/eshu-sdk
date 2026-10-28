package collector

import "os"
import "path/filepath"
import "testing"

func TestProbeSeedFormat(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x")
	if err := os.WriteFile(p, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
}
