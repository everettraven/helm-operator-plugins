package samples

type Sample interface {
	Generate()
	Prepare()
	Run()
	Path() string
}
