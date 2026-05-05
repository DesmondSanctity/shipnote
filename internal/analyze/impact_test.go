package analyze

import (
	"reflect"
	"testing"
)

func TestSinglePackageMapperAlwaysReturnsRoot(t *testing.T) {
	m := SinglePackageMapper{RootID: "myapp"}
	for _, p := range []string{"a.go", "deep/nested/x.go", ""} {
		if got := m.PackageFor(p); got != "myapp" {
			t.Errorf("PackageFor(%q) = %q, want myapp", p, got)
		}
	}
}

func TestPackagesForFilesDeduplicatedAndSorted(t *testing.T) {
	m := SinglePackageMapper{RootID: "root"}
	got := PackagesForFiles(m, []string{"a.go", "b.go", "c.go"})
	if !reflect.DeepEqual(got, []string{"root"}) {
		t.Errorf("got %+v, want [root]", got)
	}
}

func TestPackagesForFilesEmpty(t *testing.T) {
	m := SinglePackageMapper{}
	if got := PackagesForFiles(m, nil); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestPackagesForFilesNilMapper(t *testing.T) {
	if got := PackagesForFiles(nil, []string{"x"}); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}
