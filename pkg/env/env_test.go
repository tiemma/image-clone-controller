package env

import (
	"os"
	"reflect"
	"testing"
)

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	os.Exit(code)
}

func TestIsSkippableNamespace(t *testing.T) {
	if IsSkippableNamespace("", "kube-system") == false {
		t.Errorf("should be true")
	}
}

func TestGetSkippableNamespace(t *testing.T) {
	skippableNamespaces := getSkippableNamespaces()
	if !reflect.DeepEqual(skippableNamespaces, defaultNamespacesToSkip) {
		t.Errorf("should be equal, expected %s, got %s", defaultNamespacesToSkip, skippableNamespaces)
	}
}

func TestSplitCommaSeparatedString(t *testing.T) {
	specs := []struct {
		str      string
		expected []string
	}{
		{str: "default", expected: []string{"default"}},
		{str: "default,", expected: []string{"default"}},
		{str: "default, kube-system", expected: []string{"default", "kube-system"}},
		{str: "default,kube-system", expected: []string{"default", "kube-system"}},
		{str: "default,   kube-system ,", expected: []string{"default", "kube-system"}},
	}

	for _, spec := range specs {
		res := splitCommaSeparatedString(spec.str)
		if !reflect.DeepEqual(res, spec.expected) {
			t.Errorf("expected %s, got %s", res, spec.expected)
		}
	}
}
