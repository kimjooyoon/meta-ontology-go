package provenance

import (
	"errors"
	"sync"
)

// storageFaultPoint names a filesystem barrier. It is intentionally private:
// fault injection is a package test seam, not part of the storage API.
type storageFaultPoint string

const (
	faultLedgerAppendWrite         storageFaultPoint = "ledger-append-write"
	faultLedgerAppendSync          storageFaultPoint = "ledger-append-sync"
	faultLedgerAppendClose         storageFaultPoint = "ledger-append-close"
	faultPreparedWrite             storageFaultPoint = "prepared-write"
	faultPreparedSync              storageFaultPoint = "prepared-sync"
	faultPreparedClose             storageFaultPoint = "prepared-close"
	faultPreparedRename            storageFaultPoint = "prepared-rename"
	faultPreparedDirectorySync     storageFaultPoint = "prepared-directory-sync"
	faultCommittedWrite            storageFaultPoint = "committed-write"
	faultCommittedSync             storageFaultPoint = "committed-sync"
	faultCommittedClose            storageFaultPoint = "committed-close"
	faultCommittedRename           storageFaultPoint = "committed-rename"
	faultCommittedDirectorySync    storageFaultPoint = "committed-directory-sync"
	faultRecoveryLedgerWrite       storageFaultPoint = "recovery-ledger-write"
	faultRecoveryLedgerSync        storageFaultPoint = "recovery-ledger-sync"
	faultRecoveryLedgerClose       storageFaultPoint = "recovery-ledger-close"
	faultRecoveryLedgerRename      storageFaultPoint = "recovery-ledger-rename"
	faultRecoveryDirectorySync     storageFaultPoint = "recovery-directory-sync"
	faultRecoveryManifestWrite     storageFaultPoint = "recovery-manifest-write"
	faultRecoveryManifestSync      storageFaultPoint = "recovery-manifest-sync"
	faultRecoveryManifestClose     storageFaultPoint = "recovery-manifest-close"
	faultRecoveryManifestRename    storageFaultPoint = "recovery-manifest-rename"
	faultRecoveryManifestDirectory storageFaultPoint = "recovery-manifest-directory-sync"
	faultCommittedRepairCreate     storageFaultPoint = "committed-repair-create"
	faultCommittedRepairWrite      storageFaultPoint = "committed-repair-write"
	faultCommittedRepairSync       storageFaultPoint = "committed-repair-sync"
	faultCommittedRepairClose      storageFaultPoint = "committed-repair-close"
	faultCommittedRepairRename     storageFaultPoint = "committed-repair-rename"
	faultCommittedRepairDirectory  storageFaultPoint = "committed-repair-directory-sync"
	faultCommittedRepairRevalidate storageFaultPoint = "committed-repair-revalidate"
)

type storageFault struct {
	point   storageFaultPoint
	partial int
	err     error
}

var storageFaultState struct {
	sync.Mutex
	fault *storageFault
}

// installStorageFaultForTest installs one deterministic, one-shot failure.
// The returned function restores the previous test state.
func installStorageFaultForTest(point storageFaultPoint, partial int) func() {
	storageFaultState.Lock()
	previous := storageFaultState.fault
	storageFaultState.fault = &storageFault{point: point, partial: partial, err: errors.New("injected provenance filesystem failure")}
	storageFaultState.Unlock()
	return func() {
		storageFaultState.Lock()
		storageFaultState.fault = previous
		storageFaultState.Unlock()
	}
}
