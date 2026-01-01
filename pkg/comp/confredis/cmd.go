package confredis

func Command(name string, args ...any) *Cmd {
	return &Cmd{
		Name: name,
		Args: args,
	}
}

type Cmd struct {
	Name string
	Args []any
}
