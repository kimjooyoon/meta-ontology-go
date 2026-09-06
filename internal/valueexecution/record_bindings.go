package valueexecution

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func compileRecordBindings(ir semantic.IR, programs map[string]recordProgram) (map[string]string, error) {
	names := map[string]string{}
	for name, program := range programs {
		names[program.authority.activityID] = name
	}
	incoming := map[string]string{}
	for _, binding := range ir.RuntimeBindings {
		producer, producerOK := names[binding.ProducerActivity.String()]
		consumer, consumerOK := names[binding.ConsumerActivity.String()]
		if !producerOK || !consumerOK || binding.ProducerPort != semantic.RuntimeOutputPort ||
			binding.ConsumerPort != semantic.RuntimeInputPort {
			return nil, failAt(ReasonPlanInvalid, "PLAN", "resolve-record-binding", "binding does not connect compiled record activities")
		}
		if _, duplicate := incoming[consumer]; duplicate {
			return nil, failAt(ReasonPlanInvalid, "PLAN", "reject-record-input-conflict", consumer)
		}
		entityID := binding.Entity.String()
		if entityID != programs[producer].authority.outputEntityID || entityID != programs[consumer].authority.outputEntityID {
			return nil, failAt(ReasonBindingResultInvalid, "TYPECHECK", "bind-record-entity", consumer)
		}
		incoming[consumer] = producer
	}
	return incoming, nil
}

func recordExecutionOrder(programs map[string]recordProgram, incoming map[string]string) ([]string, error) {
	ready := []string{}
	outgoing := map[string][]string{}
	for name := range programs {
		if producer, bound := incoming[name]; bound {
			outgoing[producer] = append(outgoing[producer], name)
		} else {
			ready = append(ready, name)
		}
	}
	order := make([]string, 0, len(programs))
	for len(ready) > 0 {
		slices.Sort(ready)
		name := ready[0]
		ready = ready[1:]
		order = append(order, name)
		ready = append(ready, outgoing[name]...)
	}
	if len(order) != len(programs) {
		return nil, failAt(ReasonPlanInvalid, "PLAN", "reject-record-binding-cycle", "record bindings are cyclic")
	}
	return order, nil
}

func (plan RecordPlan) validateAuthority() error {
	if !validDigest(plan.sourceDigest) || plan.SourceDigest != plan.sourceDigest ||
		plan.fingerprint == "" || plan.SemanticFingerprint != plan.fingerprint ||
		len(plan.programs) == 0 || len(plan.order) != len(plan.programs) {
		return failAt(ReasonPlanInvalid, "PLAN", "validate-record-plan-authority", "plan is not CompileRecordPlan-issued")
	}
	for name, program := range plan.programs {
		if !program.authority.valid() || program.authority.activityName != name ||
			program.authority.sourceDigest != plan.sourceDigest || program.authority.semanticFingerprint != plan.fingerprint {
			return failAt(ReasonProgramAuthorityInvalid, "PLAN", "validate-record-program-authority", name)
		}
	}
	return nil
}
