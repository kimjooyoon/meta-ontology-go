package metricprogram

import (
	"path"
	"testing"
	"testing/fstest"
)

func registryFilesFixture(t *testing.T) fstest.MapFS {
	t.Helper()
	origin := RegistrySourceOrigin()
	files := fstest.MapFS{
		origin.SourcePath: {Data: registrySource}, origin.ProjectionPath: {Data: registryProjection},
	}
	entries, err := registryParts.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		raw, err := registryParts.ReadFile(entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		files[path.Join(path.Dir(origin.SourcePath), entry.Name())] = &fstest.MapFile{Data: raw}
	}
	return files
}

func TestCompiledRegistryBundleFresh1665(t *testing.T) {
	if err := VerifyCompiledRegistry(); err != nil {
		t.Fatal(err)
	}
}

func TestRegistryProjectionBundleFresh1661(t *testing.T) {
	origin := RegistrySourceOrigin()
	t.Run("complete-bundle", func(t *testing.T) {
		if err := VerifyRegistrySource(registryFilesFixture(t), origin); err != nil {
			t.Fatal(err)
		}
	})
	for name := range registryFilesFixture(t) {
		if path.Ext(name) != ".go" {
			continue
		}
		for _, operation := range []string{"omit", "change"} {
			t.Run(operation+"/"+path.Base(name), func(t *testing.T) {
				files := registryFilesFixture(t)
				if operation == "omit" {
					delete(files, name)
				} else {
					files[name] = &fstest.MapFile{Data: []byte("changed part")}
				}
				if err := VerifyRegistrySource(files, origin); err == nil {
					t.Fatal("incomplete projection bundle accepted")
				}
			})
		}
	}
	t.Run("unlisted-part", func(t *testing.T) {
		files := registryFilesFixture(t)
		files[path.Join(path.Dir(origin.SourcePath), "registry_unlisted.generated.go")] = &fstest.MapFile{}
		if err := VerifyRegistrySource(files, origin); err == nil {
			t.Fatal("unlisted projection member accepted")
		}
	})
}
