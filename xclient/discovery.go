package xclient

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type Discovery interface {
	Refresh() error
	Update(services []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}
