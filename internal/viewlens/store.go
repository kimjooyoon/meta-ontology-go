package viewlens

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/originlenspolicy"
)

// Load is explicit detached I/O. Select and the compiler never call it.
func Load(directory, actor, audience string) StoreResult {
	result := StoreResult{Schema: "gooo/view-lens-store/v1", State: "UNAVAILABLE",
		Snapshot: initialSnapshot(actor, audience)}
	path, err := profilePath(directory, actor, audience)
	if err == nil {
		var profile Profile
		profile, _, err = readProfile(path, actor, audience)
		if err == nil {
			result.Snapshot, result.State = profile.Snapshot, "LOADED"
			if profile.Snapshot.Revision == 0 {
				result.State = "ABSENT"
			}
		}
	}
	if err != nil {
		result.Issue = err.Error()
	}
	return result
}

// Record persists an event after rendering. It never waits for a busy writer.
// A stale snapshot is rejected; the last accepted event is idempotent.
func Record(directory string, raw []byte) StoreResult {
	result := StoreResult{Schema: "gooo/view-lens-store/v1", State: "UNAVAILABLE"}
	event, canonical, err := decodeEvent(raw)
	if err != nil {
		result.Issue = err.Error()
		return result
	}
	result.EventDigest = digest(canonical)
	path, err := profilePath(directory, event.Request.ActorID, event.Request.Audience)
	if err == nil {
		err = withProfileLock(path, func() error {
			return recordLocked(path, directory, event, canonical, &result)
		})
	}
	if err != nil {
		result.Issue = err.Error()
	}
	return result
}

func decodeEvent(raw []byte) (Event, []byte, error) {
	var event Event
	if err := decodeExact(raw, &event, 8192); err != nil {
		return event, nil, err
	}
	if err := originlenspolicy.VerifyProjection(); err != nil {
		return event, nil, fmt.Errorf("HINT_POLICY_UNAVAILABLE")
	}
	if err := validateRequest(event.Request); err != nil {
		return event, nil, err
	}
	if err := validateSnapshot(event.Snapshot, event.Request.ActorID, event.Request.Audience); err != nil {
		return event, nil, err
	}
	if event.Schema != "gooo/view-lens-event/v1" || event.Choice != choose(event.Snapshot, event.Request) {
		return event, nil, fmt.Errorf("HINT_EVENT_ORIGIN_MISMATCH")
	}
	canonical, err := json.Marshal(event)
	return event, canonical, err
}

func profilePath(directory, actor, audience string) (string, error) {
	if err := validateIdentity(actor, audience); err != nil {
		return "", err
	}
	if directory == "" {
		return "", fmt.Errorf("PROFILE_DIRECTORY_REQUIRED")
	}
	key := digest([]byte(Domain + "\x00" + actor + "\x00" + audience))[7:]
	return filepath.Join(directory, key+".gopack"), nil
}

func recordLocked(path, directory string, event Event, raw []byte, result *StoreResult) error {
	profile, priorRaw, err := readProfile(path, event.Request.ActorID, event.Request.Audience)
	if err != nil {
		return err
	}
	result.Snapshot = profile.Snapshot
	archive := filepath.Join(directory, "events", result.EventDigest[7:]+".gopack")
	if profile.LastEventDigest == result.EventDigest {
		stored, _, err := readPacked(archive)
		if err != nil || !bytes.Equal(stored, raw) {
			return fmt.Errorf("PROFILE_ARCHIVE_UNAVAILABLE")
		}
		result.State, result.Persisted, result.ArchivePath = "ALREADY_RECORDED", true, archive
		return nil
	}
	if profile.Snapshot != event.Snapshot {
		return fmt.Errorf("PROFILE_SNAPSHOT_STALE")
	}
	if len(priorRaw) > 0 {
		previous := filepath.Join(directory, "profiles", digest(priorRaw)[7:]+".gopack")
		if _, err := archiveEvent(previous, priorRaw); err != nil {
			return err
		}
	}
	if err := advanceProfile(&profile, event, result.EventDigest); err != nil {
		return err
	}
	n, err := archiveEvent(archive, raw)
	if err != nil {
		return err
	}
	next, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	profileArchive := filepath.Join(directory, "profiles", digest(next)[7:]+".gopack")
	if _, err := archiveEvent(profileArchive, next); err != nil {
		return err
	}
	if _, err := writeProfile(path, next); err != nil {
		return err
	}
	profile.Snapshot.ProfileDigest = digest(next)
	result.State, result.Persisted, result.ArchivePath = "SAVED", true, archive
	result.ProfileArchivePath = profileArchive
	result.Snapshot, result.RawBytes, result.StoredBytes = profile.Snapshot, len(raw), n
	return nil
}

func advanceProfile(profile *Profile, event Event, eventDigest string) error {
	if profile.Snapshot.Revision >= countLimit {
		return fmt.Errorf("PROFILE_SELECTION_COUNT_BOUND")
	}
	if event.Request.Prefer != "" {
		profile.Snapshot.Preferred, _ = ResolveView(event.Request.Prefer)
	}
	profile.Snapshot.Revision++
	if event.Choice.ViewID == Full {
		profile.Snapshot.FullSelections++
	} else {
		profile.Snapshot.NamedSelections++
	}
	profile.Snapshot.ProfileDigest = ""
	profile.LastEventDigest = eventDigest
	profile.Recent = append(profile.Recent, event.Choice)
	if len(profile.Recent) > historyLimit {
		copy(profile.Recent, profile.Recent[len(profile.Recent)-historyLimit:])
		profile.Recent = profile.Recent[:historyLimit]
	}
	return nil
}
