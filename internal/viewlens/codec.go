package viewlens

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/originpack"
)

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}

func decodeExact(raw []byte, value any, limit int) error {
	if len(raw) > limit {
		return fmt.Errorf("HINT_BYTE_BOUND")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("HINT_SYNTAX_INVALID")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("HINT_TRAILING_DATA")
	}
	return nil
}

func decodeSnapshot(raw []byte, actor, audience string) (Snapshot, error) {
	var snapshot Snapshot
	if err := decodeExact(raw, &snapshot, hintLimit); err != nil {
		return Snapshot{}, err
	}
	if err := validateSnapshot(snapshot, actor, audience); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func validateSnapshot(snapshot Snapshot, actor, audience string) error {
	if snapshot.Schema != "gooo/view-lens-hint/v1" || snapshot.Domain != Domain ||
		snapshot.ActorID != actor || snapshot.Audience != audience ||
		snapshot.Revision < 0 || snapshot.Revision > countLimit ||
		snapshot.FullSelections < 0 || snapshot.NamedSelections < 0 ||
		snapshot.FullSelections > countLimit || snapshot.NamedSelections > countLimit ||
		snapshot.FullSelections+snapshot.NamedSelections != snapshot.Revision {
		return fmt.Errorf("HINT_IDENTITY_OR_COUNT_MISMATCH")
	}
	if snapshot.Preferred != "" && snapshot.Preferred != Full && snapshot.Preferred != NamedInputs {
		return fmt.Errorf("HINT_VIEW_UNAVAILABLE")
	}
	if (snapshot.Revision == 0 && snapshot.ProfileDigest != "") ||
		(snapshot.Revision > 0 && !validDigest(snapshot.ProfileDigest)) {
		return fmt.Errorf("HINT_PROFILE_ORIGIN_INVALID")
	}
	return nil
}

func readPacked(path string) ([]byte, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()
	packed, err := io.ReadAll(io.LimitReader(file, originpack.MaxRawBytes+46))
	if err != nil {
		return nil, len(packed), err
	}
	raw, err := originpack.Unpack(packed)
	return raw, len(packed), err
}

func readProfile(path, actor, audience string) (Profile, []byte, error) {
	raw, _, err := readPacked(path)
	if os.IsNotExist(err) {
		return Profile{Schema: Schema, Snapshot: initialSnapshot(actor, audience), Recent: []Choice{}}, nil, nil
	}
	if err != nil {
		return Profile{}, raw, err
	}
	var profile Profile
	if err := decodeExact(raw, &profile, originpack.MaxRawBytes); err != nil {
		return Profile{}, raw, err
	}
	canonical, err := json.Marshal(profile)
	if err != nil || !bytes.Equal(raw, canonical) {
		return Profile{}, raw, fmt.Errorf("PROFILE_CANONICAL_FORMAT_REQUIRED")
	}
	if profile.Schema != Schema || profile.Snapshot.ProfileDigest != "" || profile.Snapshot.Revision == 0 ||
		!validDigest(profile.LastEventDigest) || int64(len(profile.Recent)) != min(profile.Snapshot.Revision, historyLimit) {
		return Profile{}, raw, fmt.Errorf("PROFILE_HISTORY_INVALID")
	}
	for _, choice := range profile.Recent {
		if (choice.ViewID != Full && choice.ViewID != NamedInputs) || !boundedIdentity(choice.ActivityID, 512) ||
			!validDigest(choice.SourceDigest) || !boundedIdentity(choice.Reason, 128) ||
			!validDigest(choice.Policy.SourceDigest) || !validDigest(choice.Policy.ProjectionDigest) {
			return Profile{}, raw, fmt.Errorf("PROFILE_HISTORY_ORIGIN_INVALID")
		}
	}
	profile.Snapshot.ProfileDigest = digest(raw)
	if err := validateSnapshot(profile.Snapshot, actor, audience); err != nil {
		return Profile{}, raw, err
	}
	return profile, raw, nil
}
