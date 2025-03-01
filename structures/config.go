package structures

type Config []Target
type Target struct {
	Bot     uint64 `yaml:"bot"`
	Channel uint64 `yaml:"channel"`
	Keyword string `yaml:"keyword"`
	Timeout int    `yaml:"timeout"`
}
