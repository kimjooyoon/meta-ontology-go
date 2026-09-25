package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveConfig(settings config) (config, error) {
	var err error
	settings.root, err = filepath.Abs(settings.root)
	if err != nil {
		return config{}, err
	}
	settings.physical, err = filepath.Abs(settings.physical)
	if err != nil {
		return config{}, err
	}
	settings.work, err = filepath.Abs(settings.work)
	if err != nil {
		return config{}, err
	}
	if settings.index == "" {
		settings.index = filepath.Join(settings.work, "logical.index")
	} else {
		settings.index, err = filepath.Abs(settings.index)
	}
	return settings, err
}

func prepareWork(settings config) error {
	if settings.work == "" || settings.work == string(filepath.Separator) {
		return fmt.Errorf("safe empty work directory is required")
	}
	relative, err := filepath.Rel(settings.root, settings.work)
	if err != nil {
		return err
	}
	if relative == "." || !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("work directory must be outside the repository")
	}
	if _, err := os.Stat(settings.work); !os.IsNotExist(err) {
		return fmt.Errorf("work directory must not exist: %s", settings.work)
	}
	return os.MkdirAll(settings.work, 0o755)
}

func resolvePath(root, name string) (string, error) {
	if name == "" || filepath.IsAbs(name) || strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("unsafe projected path %q", name)
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe projected path %q", name)
	}
	return filepath.Join(root, clean), nil
}

func samePath(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}
