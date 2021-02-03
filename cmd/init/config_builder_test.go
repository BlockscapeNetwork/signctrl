package init

import (
	"os"
	"testing"
)

func TestBuildConfigTemplate(t *testing.T) {
	defer os.Remove("./pairmint.toml")
	if err := BuildConfigTemplate("."); err != nil {
		t.Errorf("Expected err to be nil, instead got: %v", err)
	}
	if err := BuildConfigTemplate("/some/non/existent/path"); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
}
