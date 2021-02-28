package docker

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	repoURL = "docker.io/kube456"

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func TestGetReference(t *testing.T) {
	specs := []struct {
		img      string
		expected string
	}{
		{img: "docker.io/kube123/test:123", expected: "docker.io/kube123/test:123"},
		{img: "docker.io/kube123/test", expected: "docker.io/kube123/test:latest"},
	}

	for _, spec := range specs {
		ref, err := getReference(spec.img)
		if err != nil {
			t.Errorf("error occured getting reference: %s", err)
		}
		if !strings.Contains(ref.Name(), spec.img) {
			t.Errorf("expected %s, got %s", ref.Name(), spec.img)
		}
	}
}

func TestGetCacheImageURL(t *testing.T) {
	specs := []struct {
		img      string
		expected string
	}{
		{img: "docker.io/kube123/test:123", expected: fmt.Sprintf("%s/test:123", repoURL)},
		{img: "docker.io/kube123/test", expected: fmt.Sprintf("%s/test:latest", repoURL)},
	}

	for _, spec := range specs {
		ref, err := getReference(spec.img)
		if err != nil {
			t.Errorf("error occured getting reference: %s", err)
		}

		res := getCacheImageURL(ref)
		if res != spec.expected {
			t.Errorf("expected %s, got %s", spec.expected, res)
		}
	}
}
