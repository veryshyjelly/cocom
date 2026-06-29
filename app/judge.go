package app

type Problem struct {
	Title       string
	Url         string
	MemoryLimit uint64
	TimeLimit   uint64
}

type Testcase struct {
	Input  string
	Output string
	Error  string
	Answer string
	Status string
	Time   float64
}

func (m Model) Compile() error {
	return nil
}

func (m Model) Run() error {
	return nil
}
