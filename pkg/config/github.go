package config

type GithubUser struct {
	Name string `yaml:"name"`
}

type GithubCrawlConfig struct {
	BotAddr string                `yaml:"bot-addr"`
	Repos   []string              `yaml:"repos"`
	Users   map[string]GithubUser `yaml:"users"`
}
