package toolchainrelease

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strconv"
)

func verifyArchiveMembers(path string, target Target, members []ArchiveMemberEvidence) error {
	if len(members) != 5 {
		return fmt.Errorf("TOOLCHAIN_RELEASE_MEMBER_INVENTORY_MISMATCH")
	}
	want := make(map[string]ArchiveMemberEvidence, len(members))
	for _, member := range members {
		limit := int64(64 << 10)
		if member.Role == "binary" {
			limit = 128 << 20
		}
		if member.Name == "" || want[member.Name].Name != "" || member.Bytes < 1 || member.Bytes > limit {
			return fmt.Errorf("TOOLCHAIN_RELEASE_MEMBER_INVENTORY_MISMATCH")
		}
		want[member.Name] = member
	}
	switch target.ArchiveFormat {
	case "tar.gz":
		return verifyTarMembers(path, want)
	case "zip":
		return verifyZipMembers(path, want)
	default:
		return fmt.Errorf("TOOLCHAIN_RELEASE_ARCHIVE_FORMAT_UNKNOWN")
	}
}

func verifyTarMembers(path string, want map[string]ArchiveMemberEvidence) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	seen := map[string]bool{}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		member, exists := want[header.Name]
		mode, _ := strconv.ParseInt(member.Mode, 8, 64)
		if !exists || seen[header.Name] || len(seen) >= len(want) ||
			(header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA) ||
			header.Mode != mode || header.Size != member.Bytes {
			return fmt.Errorf("TOOLCHAIN_RELEASE_MEMBER_INVENTORY_MISMATCH")
		}
		if err := verifyMemberBytes(reader, member); err != nil {
			return err
		}
		seen[header.Name] = true
	}
	if len(seen) != len(want) {
		return fmt.Errorf("TOOLCHAIN_RELEASE_MEMBER_INVENTORY_MISMATCH")
	}
	return nil
}

func verifyZipMembers(path string, want map[string]ArchiveMemberEvidence) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer reader.Close()
	if len(reader.File) != len(want) {
		return fmt.Errorf("TOOLCHAIN_RELEASE_MEMBER_INVENTORY_MISMATCH")
	}
	seen := map[string]bool{}
	for _, file := range reader.File {
		member, exists := want[file.Name]
		mode, _ := strconv.ParseUint(member.Mode, 8, 32)
		if !exists || seen[file.Name] || !file.Mode().IsRegular() ||
			file.Mode().Perm() != os.FileMode(mode) || file.UncompressedSize64 != uint64(member.Bytes) {
			return fmt.Errorf("TOOLCHAIN_RELEASE_MEMBER_INVENTORY_MISMATCH")
		}
		content, err := file.Open()
		if err != nil {
			return err
		}
		check := verifyMemberBytes(content, member)
		closeErr := content.Close()
		if check != nil {
			return check
		}
		if closeErr != nil {
			return closeErr
		}
		seen[file.Name] = true
	}
	return nil
}

func verifyMemberBytes(reader io.Reader, member ArchiveMemberEvidence) error {
	hash := sha256.New()
	count, err := io.Copy(hash, io.LimitReader(reader, member.Bytes+1))
	if err != nil || count != member.Bytes || fmt.Sprintf("sha256:%x", hash.Sum(nil)) != member.Digest {
		return fmt.Errorf("TOOLCHAIN_RELEASE_MEMBER_CONTENT_MISMATCH: %s", member.Name)
	}
	return nil
}
