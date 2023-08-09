package types

func NewEventSlashedBTCValidator(evidence *Evidence) *EventSlashedBTCValidator {
	return &EventSlashedBTCValidator{
		Evidence: evidence,
	}
}
