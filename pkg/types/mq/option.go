package mq

type Option interface {
	OptionScheme() string
}

type OptionApplier interface {
	Apply(Option)
}

type OptionApplyFunc func(Option)

func (f OptionApplyFunc) Apply(opt Option) { f(opt) }
