package modelschemas

type EventStatus string

const (
	EventStatusPending EventStatus = "pending"
	EventStatusSuccess EventStatus = "success"
	EventStatusFailed  EventStatus = "failed"
)

func (e EventStatus) Ptr() *EventStatus {
	return &e
}
