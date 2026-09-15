package workfrontier

func (p *RepairPath) UnmarshalJSON(data []byte) error {
	type wire RepairPath
	var raw wire
	fields, err := decodeStrictObject(data, &raw)
	if err != nil {
		return err
	}
	*p = RepairPath(raw)
	p.stableIDPresent = present(fields, "stable_id")
	p.obligationIDPresent = present(fields, "obligation_id")
	p.prerequisiteObligationIDsPresent = present(fields, "prerequisite_obligation_ids")
	p.readSetPresent = present(fields, "read_set")
	p.writeSetPresent = present(fields, "write_set")
	p.requiredPressureIDsPresent = present(fields, "required_pressure_ids")
	p.policyPriorityPresent = present(fields, "policy_priority")
	p.cpuCoreNSUpperBoundPresent = present(fields, "cpu_core_ns_upper_bound")
	p.fromJSON = true
	return nil
}
func (c *Capacity) UnmarshalJSON(data []byte) error {
	type wire Capacity
	var raw wire
	fields, err := decodeStrictObject(data, &raw)
	if err != nil {
		return err
	}
	*c = Capacity(raw)
	c.cpuCoreNSPresent = present(fields, "cpu_core_ns")
	return nil
}
func (in *Input) UnmarshalJSON(data []byte) error {
	type wire Input
	var raw wire
	fields, err := decodeStrictObject(data, &raw)
	if err != nil {
		return err
	}
	*in = Input(raw)
	in.fromJSON = true
	in.present = inputPresence{
		schemaVersion:            present(fields, "schema_version"),
		snapshotDigest:           present(fields, "snapshot_digest"),
		policyDigest:             present(fields, "policy_digest"),
		registryDigest:           present(fields, "registry_digest"),
		minimumSelectedPressures: present(fields, "minimum_selected_pressures"),
		capacity:                 present(fields, "capacity"),
		pressures:                present(fields, "pressures"),
		states:                   present(fields, "states"),
		paths:                    present(fields, "paths"),
	}
	return nil
}
