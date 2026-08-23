package languagepackageruntime

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"

func baseManifest() packageruntime.Manifest {
	return packageruntime.Manifest{
		Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Run"},
		Packages: []packageruntime.PackageSpec{
			packageSpec("example/core", "core", nil, source("core.gooo", `package core
namespace core
entity Core id "gooo://runtime/core"`)),
			packageSpec("example/left", "left", []string{"example/core"}, source("left.gooo", `package left
namespace left
entity Left id "gooo://runtime/left"`)),
			packageSpec("example/right", "right", []string{"example/core"}, source("right.gooo", `package right
namespace right
entity Right id "gooo://runtime/right"`)),
			packageSpec("example/app", "app", []string{"example/right", "example/left", "example/left"},
				source("main.gooo", `package app
namespace app
entity Request id "gooo://runtime/request"
activity Run(Request) -> Request`),
				source("extra.gooo", `package app
namespace app
entity Extra id "gooo://runtime/extra"`)),
		},
	}
}

func packageSpec(path, name string, imports []string, sources ...packageruntime.Source) packageruntime.PackageSpec {
	return packageruntime.PackageSpec{Path: path, Name: name, Imports: imports, Sources: sources}
}

func source(filename, content string) packageruntime.Source {
	return packageruntime.Source{Filename: filename, Content: content}
}
