package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type SetterConfig struct {
	GithubToken        string                      `yaml:"GithubToken"`
	SchemesMasterURL   string                      `yaml:"SchemesMasterURL"`
	TemplatesMasterURL string                      `yaml:"TemplatesMasterURL"`
	SchemesListFile    string                      `yaml:"SchemesListFile"`
	TemplatesListFile  string                      `yaml:"TemplatesListFile"`
	SchemesCachePath   string                      `yaml:"SchemesCachePath"`
	TemplatesCachePath string                      `yaml:"TemplatesCachePath"`
	DryRun             bool                        `yaml:"DryRun"`
	Colorscheme        string                      `yaml:"Colorscheme"`
	Applications       map[string]StetterAppConfig `yaml:"applications"`
}

type StetterAppConfig struct {
	Enabled bool              `yaml:"enabled"`
	Hook    string            `yaml:"hook"`
	Mode    string            `yaml:"mode"`
	Prefix  string            `yaml:"prefix"`
	Files   map[string]string `yaml:"files"`
}

func NewConfig(path string) SetterConfig {
	var conf SetterConfig

  conf.GithubToken = ""
  conf.SchemesMasterURL = "https://raw.githubusercontent.com/chriskempson/base16-schemes-source/master/list.yaml"
  conf.TemplatesMasterURL = "https://raw.githubusercontent.com/chriskempson/base16-templates-source/master/list.yaml"
  conf.SchemesListFile = "cache/schemeslist.yaml"
  conf.TemplatesListFile = "cache/templateslist.yaml"
  conf.SchemesCachePath = "cache/schemes/"
  conf.TemplatesCachePath = "cache/templates/"
  conf.DryRun = true
  conf.Colorscheme = "flat.yaml"
  conf.Applications = map[string]StetterAppConfig{}

  file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read config. Using defaults: %v\n", err)
	} else {
	  check(err)
	  err = yaml.Unmarshal((file), &conf)
	  check(err)
    for k := range conf.Applications {
      if conf.Applications[k].Mode == "" {
        app := conf.Applications[k]
        app.Mode = "rewrite"
        conf.Applications[k] = app
      }
      if conf.Applications[k].Prefix == "" {
        app := conf.Applications[k]
        app.Prefix = "#"
        conf.Applications[k] = app
      }
    }

  }
	return conf
}
func (c SetterConfig) Show() {
	fmt.Println("GithubToken: ", c.GithubToken)
	fmt.Println("SchemesListFile: ", c.SchemesListFile)
	fmt.Println("TemplatesListFile: ", c.TemplatesListFile)
	fmt.Println("SchemesCachePath: ", c.SchemesCachePath)
	fmt.Println("TemplatesCachePath: ", c.TemplatesCachePath)
	fmt.Println("DryRun: ", c.DryRun)

	for k, v := range c.Applications {
		fmt.Println("  App: ", k)
		fmt.Println("    Enabled: ", v.Enabled)
    fmt.Println("    Mode: ", v.Mode)
		fmt.Println("    Hook: ", v.Hook)
		fmt.Println("    Prefix: ", v.Prefix)
		for k1, v1 := range v.Files {
			fmt.Println("      ", k1, "  ", v1)
		}
	}
}

type Application1 struct {
}
