package worker

// NewStatus fn
func NewStatus() *Status {
	return &Status{Workers: NewWorkers()}
}

// Status struct
type Status struct {
	*Workers
}

// Len fn
func (s *Status) Len() int {
	return len(s.List())
}

// List fn
func (s *Status) List() map[string]*worker {
	return s.Channels()
}
