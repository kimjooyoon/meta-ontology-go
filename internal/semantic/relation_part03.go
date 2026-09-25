package semantic

import (
	"fmt"
	"strings"
)

func NewCandidateFact(subject ID, predicate Relation, object ID, reason string) Fact {
	return Fact{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
		Status:    FactCandidate,
		Reason:    strings.TrimSpace(reason),
	}
}
func NewUsedFact(activity, entity ID) Fact {
	return NewFact(activity, Used, entity)
}
func NewWasGeneratedByFact(entity, activity ID) Fact {
	return NewFact(entity, WasGeneratedBy, activity)
}
func NewWasDerivedFromFact(entity, source ID) Fact {
	return NewFact(entity, WasDerivedFrom, source)
}
func NewWasAssociatedWithFact(activity, agent ID) Fact {
	return NewFact(activity, WasAssociatedWith, agent)
}
func (f Fact) Key() FactKey {
	return FactKey{Subject: f.Subject, Predicate: f.Predicate, Object: f.Object}
}
func (f Fact) Normalized() (Fact, error) {
	subject, err := ParseIdentity(f.Subject.String())
	if err != nil {
		return Fact{}, fmt.Errorf("%w: subject: %v", ErrInvalidFact, err)
	}
	object, err := ParseIdentity(f.Object.String())
	if err != nil {
		return Fact{}, fmt.Errorf("%w: object: %v", ErrInvalidFact, err)
	}
	if !f.Predicate.Valid() {
		return Fact{}, fmt.Errorf("%w: %v", ErrUnknownRelation, f.Predicate)
	}
	if f.Status == 0 {

		f.Status = FactDeterministic
	}
	if f.Status != FactDeterministic && f.Status != FactCandidate {
		return Fact{}, fmt.Errorf("%w: unknown status %d", ErrInvalidFact, f.Status)
	}
	span := f.Span.Normalized()
	if err := span.Validate(); err != nil {
		return Fact{}, fmt.Errorf("%w: span: %v", ErrInvalidFact, err)
	}

	f.Subject = subject
	f.Object = object
	f.Span = span
	f.Reason = strings.TrimSpace(f.Reason)
	return f, nil
}
func (f Fact) Validate() error {
	_, err := f.Normalized()
	return err
}
func (f Fact) WithSpan(span Span) Fact {
	f.Span = span
	return f
}
func (f Fact) WithReason(reason string) Fact {
	f.Reason = strings.TrimSpace(reason)
	return f
}
