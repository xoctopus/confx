package confmq

type Option interface {
	Scheme() string
}

type OptionApplier interface {
	Apply(Option)
}

type OptionApplyFunc func(Option)

func (f OptionApplyFunc) Apply(opt Option) { f(opt) }
