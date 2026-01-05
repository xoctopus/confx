package confmq

type Option interface {
	Scheme() string
	Role() Role
}

type OptionApplier interface {
	Apply(Option)
}

type OptionApplyFunc func(Option)

func (f OptionApplyFunc) Apply(opt Option) { f(opt) }

type Role int

const (
	Publisher Role = 1
	Consumer  Role = 2
)
