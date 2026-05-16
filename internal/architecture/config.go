package architecture

type Config struct {
	Layers   []Layer  `yaml:"layers"`
	Settings Settings `yaml:"settings"`
}

type Layer struct {
	Name     string   `yaml:"name"`
	Patterns []string `yaml:"patterns"`
	Allow    []string `yaml:"allow"`
}

type Settings struct {
	AllowUnknown    *bool `yaml:"allow_unknown"`
	AllowSelf       *bool `yaml:"allow_self"`
	DefaultAllowAll *bool `yaml:"default_allow_all"`
}
