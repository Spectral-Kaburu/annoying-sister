// Package nudge defines the NudgeEvent type shared across producer
// goroutines and the TTS consumer, plus the template-based ragebait text
// generation for each nudge category.
package nudge

// Category identifies which producer generated a NudgeEvent.
type Category string

const (
	CategoryIdle           Category = "idle"
	CategoryDormantProject Category = "dormant_project"
	CategoryOnStart        Category = "on_start"
)

// Event is sent by a producer goroutine onto the shared nudge channel.
// Producers make the decision to nudge and fully render the text; the TTS
// consumer only ever writes Text to the socket, it never has category- or
// template-specific knowledge.
type Event struct {
	Category Category
	Text     string
}
