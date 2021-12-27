package modelschemas

type EventStatus string

const (
	EventStatusPending EventStatus = "pending"
	EventStatusSuccess EventStatus = "success"
	EventStatusFailure EventStatus = "failure"
)

func (e EventStatus) Ptr() *EventStatus {
	return &e
}
