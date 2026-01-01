package builtins

type DepType string

const (
	DEP_BUILD  DepType = "BUILD"
	DEP_TEST   DepType = "TEST"
	DEP_FORMAT DepType = "FORMAT"
	DEP_LINT   DepType = "LINT"
)

type Makefile struct {
}

type Dependency struct {
	// Name dependency command name
	Name string
	// Type dependency type
	Type DepType
	// Repo repository
	Repo string
	// Desc description
	Desc string
}

var requirements = []Dependency{
	{Name: "go", Type: DEP_BUILD},
	{Name: "go", Type: DEP_TEST},
}

var dependencies = []Dependency{
	{
		Name: "goimports-reviser",
		Type: DEP_FORMAT,
		Repo: "github.com/incu6us/goimports-reviser/v3",
	},
	{
		Name: "ineffassign",
		Type: DEP_LINT,
		Repo: "github.com/gordonklaus/ineffassign",
	},
	{
		Name: "gocyclo",
		Type: DEP_LINT,
	},
}
