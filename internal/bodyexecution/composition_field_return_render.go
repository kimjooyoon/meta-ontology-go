package bodyexecution

// Finish the value before copying its field slots into the return boundary.
// Go evaluates ordinary call arguments left-to-right, so passing the slot array
// beside FieldAssignRecord would otherwise capture its pre-assignment state.
func (r compositionFieldRenderer) returnWrite(n compositionFieldNode) string {
	name := r.prefix + "FieldReturnedValue"
	kind := r.plan.inventory.bindings[n.binding].Record.GoName
	return "func() " + kind + " {" + name + ":=" + r.recordWrite(n) + ";return " +
		r.call("FieldReturn", r.slot(n.binding), name) + "}()"
}
