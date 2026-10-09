package toolchainrelease

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const packageInterfaceManifest = "examples/package-optional-record-flow/gooo.workspace.json"

func smokePackageInterface(binary string, input BuildInput) error {
	return smokePackageInterfaceWithRunner(binary, input, commandOutput)
}

func smokePackageInterfaceWithRunner(binary string, input BuildInput,
	run func(string, []string, string, ...string) ([]byte, error)) error {
	var first []byte
	for _, mode := range []string{"initial", "repeated", "plain"} {
		args := []string{"package", "interface"}
		extension := ".txt"
		if mode != "plain" {
			args, extension = append(args, "--json"), ".json"
		}
		raw, runErr := run(input.Root, nil, binary, append(args, packageInterfaceManifest)...)
		if runErr != nil {
			extension = ".failed-output"
		}
		name := input.Target.ID + "-package-interface-" + mode + extension
		if err := os.WriteFile(filepath.Join(input.OutputDir, name), raw, 0o644); err != nil {
			return errors.Join(runErr, fmt.Errorf("retain package interface output: %w", err))
		}
		if runErr != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_INTERFACE %s: %w", mode, runErr)
		}
		if mode == "plain" {
			if string(raw) != packageInterfacePlain {
				return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_INTERFACE plain declarations differ")
			}
			continue
		}
		if err := validatePackageInterfaceSmoke(raw, input.Root); err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_INTERFACE %s: %w", mode, err)
		}
		if mode == "repeated" && !bytes.Equal(first, raw) {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_INTERFACE repeated projection differs")
		}
		first = bytes.Clone(raw)
	}
	return nil
}

const packageInterfacePlain = `package example/domain (domain)
  activity Submit(example://domain/profile) -> example://domain/profile id=domain://activity/submit
  record Profile id=example://domain/profile
    note: urn:gooo:type:string optional one id=example://domain/profile/note
    complete: urn:gooo:type:boolean optional one id=example://domain/profile/complete
    count: urn:gooo:type:integer optional one id=example://domain/profile/count
package example/app (app)
  activity Main(example://domain/profile) -> example://domain/profile id=app://activity/main
`
