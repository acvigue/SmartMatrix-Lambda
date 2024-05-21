package smartmatrixlambda

type AppletManifest struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Summary     string `yaml:"summary"`
	Description string `yaml:"desc"`
	Author      string `yaml:"author"`
	FileName    string `yaml:"fileName"`
	PackageName string `yaml:"packageName"`
}
