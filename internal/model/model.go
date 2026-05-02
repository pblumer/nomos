package model

type Cosmos struct {
	ID       string   `yaml:"id" json:"id"`
	Type     string   `yaml:"type" json:"type"`
	Name     string   `yaml:"name" json:"name"`
	Version  string   `yaml:"version" json:"version"`
	Status   string   `yaml:"status" json:"status"`
	Owner    string   `yaml:"owner" json:"owner"`
	Summary  string   `yaml:"summary" json:"summary"`
	Domains  []string `yaml:"domains" json:"domains"`
}

type Domain struct {
	ID       string   `yaml:"id" json:"id"`
	Type     string   `yaml:"type" json:"type"`
	Name     string   `yaml:"name" json:"name"`
	Version  string   `yaml:"version" json:"version"`
	Status   string   `yaml:"status" json:"status"`
	Owner    string   `yaml:"owner" json:"owner"`
	DNSName  string   `yaml:"dns_name" json:"dns_name"`
	Summary  string   `yaml:"summary" json:"summary"`
	Services []string `yaml:"services" json:"services"`
}

type Service struct {
	ID string `yaml:"id" json:"id"`
	Type string `yaml:"type" json:"type"`
	Name string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	Status string `yaml:"status" json:"status"`
	Owner string `yaml:"owner" json:"owner"`
	Summary string `yaml:"summary" json:"summary"`
}

type Finding struct { Severity, Code, Message, Path, Recommendation string }
