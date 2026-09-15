package generator

import (
	"fmt"
)

func normalizePorts(activity *Activity, entityTypes map[string]string) error {
	if err := normalizePortList(activity, &activity.Inputs, entityTypes, "input"); err != nil {
		return err
	}
	return normalizePortList(activity, &activity.Outputs, entityTypes, "output")
}
func normalizePortList(activity *Activity, ports *[]Port, entityTypes map[string]string, direction string) error {
	seen := make(map[string]struct{}, len(*ports))
	for index := range *ports {
		port := &(*ports)[index]
		if err := normalizePort(port, entityTypes, activity.ID, direction, index); err != nil {
			return err
		}
		if _, exists := seen[port.GoName]; exists {
			return fmt.Errorf("generator: activity %q has duplicate %s name %q", activity.ID, direction, port.GoName)
		}
		seen[port.GoName] = struct{}{}
	}
	return nil
}
func normalizePort(port *Port, entityTypes map[string]string, activityID, direction string, index int) error {
	if port.EntityID == "" {
		port.EntityID = port.ID
	}
	entityType, hasEntity := entityTypes[port.EntityID]
	if port.EntityID != "" && !hasEntity {
		return fmt.Errorf("generator: activity %q references unknown entity %q", activityID, port.EntityID)
	}
	if port.GoName == "" {
		port.GoName = port.Name
	}
	if port.GoName == "" && hasEntity {
		port.GoName = lowerCamel(entityType)
	}
	if port.GoName == "" {
		return fmt.Errorf("generator: activity %q %s %d has no parameter name", activityID, direction, index)
	}
	if !isGoIdentifier(port.GoName) {
		return fmt.Errorf("generator: activity %q has invalid %s name %q", activityID, direction, port.GoName)
	}
	if port.GoType == "" && hasEntity {
		port.GoType = entityType
	}
	if port.GoType == "" {
		return fmt.Errorf("generator: activity %q %s %q has no Go type or entity reference", activityID, direction, port.GoName)
	}
	return nil
}
func defaultActivityBody(activity *Activity, entityTypes map[string]string) string {
	if len(activity.Outputs) == 0 {
		return ""
	}
	results := make([]string, len(activity.Outputs))
	for index, output := range activity.Outputs {
		if entityType, ok := entityTypes[output.EntityID]; ok && entityType == output.GoType {

			results[index] = output.GoType + "{}"
			continue
		}

		results[index] = "*new(" + output.GoType + ")"
	}
	return "return " + join(results, ", ")
}
