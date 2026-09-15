package provenance

func takeStorageFault(point storageFaultPoint) (*storageFault, bool) {
	storageFaultState.Lock()
	defer storageFaultState.Unlock()
	if storageFaultState.fault == nil || storageFaultState.fault.point != point {
		return nil, false
	}
	fault := storageFaultState.fault
	storageFaultState.fault = nil
	return fault, true
}
func failStorageOperation(point storageFaultPoint) error {
	fault, ok := takeStorageFault(point)
	if !ok {
		return nil
	}
	return fault.err
}
