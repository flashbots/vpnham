package monitor

type Status int8

const (
	Down    Status = -1
	Pending Status = 0
	Up      Status = 1
)

func (s Status) String() string {
	switch s {
	case Down:
		return "DOWN"
	case Pending:
		return "PENDING"
	case Up:
		return "UP"
	}
	return "N/A"
}
