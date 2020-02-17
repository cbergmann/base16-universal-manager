package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"gopkg.in/yaml.v2"
)

var defaultSchemesMasterURL = "https://raw.githubusercontent.com/chriskempson/base16-schemes-source/master/list.yaml"
var defaultTemplatesMasterURL = "https://raw.githubusercontent.com/chriskempson/base16-templates-source/master/list.yaml"

// SetterConfig is the applicaton's configuration.
type SetterConfig struct {
	GithubToken        string                     `yaml:"GithubToken"`
	SchemesMasterURL   string                     `yaml:"SchemesMasterURL"`
	TemplatesMasterURL string                     `yaml:"TemplatesMasterURL"`
	DryRun             bool                       `yaml:"DryRun"`
	Colorscheme        string                     `yaml:"Colorscheme"`
	Applications       map[string]SetterAppConfig `yaml:"applications"`
	SchemesCachePath   string
	SchemesListFile    string
	TemplatesCachePath string
	TemplatesListFile  string
}

// SetterAppConfig is the configuration for a particular application being themed.
type SetterAppConfig struct {
	Enabled bool              `yaml:"enabled"`
	Hook    string            `yaml:"hook"`
	Hooks   []string          `yaml:"hooks"`
	Mode    string            `yaml:"mode"`
	Comment_Prefix  string            `yaml:"comment_prefix"`
	Files   map[string]string `yaml:"files"`
}

// NewConfig parses the provided configuration file and returns the app configuration.
func NewConfig(path string) SetterConfig {
	if path == "" {
		fmt.Fprintf(os.Stderr, "no config file found\n")
		os.Exit(1)
	}

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
  conf.Applications = map[string]SetterAppConfig{}

  file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read config. Using defaults: %v\n", err)
	} else {
	  check(err)
	  err = yaml.Unmarshal((file), &conf)
	  check(err)

    if conf.SchemesMasterURL == "" {
		  conf.SchemesMasterURL = defaultSchemesMasterURL
    }
	  if conf.TemplatesMasterURL == "" {
      conf.TemplatesMasterURL = defaultTemplatesMasterURL
    }

    conf.SchemesListFile = expandPath(conf.SchemesListFile)
    conf.TemplatesListFile = expandPath(conf.TemplatesListFile)
    conf.SchemesCachePath = expandPath(conf.SchemesCachePath)
    conf.TemplatesCachePath = expandPath(conf.TemplatesCachePath)
    for k := range conf.Applications {
      if conf.Applications[k].Mode == "" {
        app := conf.Applications[k]
        app.Mode = "rewrite"
        conf.Applications[k] = app
      }
      if conf.Applications[k].Comment_Prefix == "" {
        app := conf.Applications[k]
        app.Comment_Prefix = "#"
        conf.Applications[k] = app
      }
      for f := range conf.Applications[k].Files {
        conf.Applications[k].Files[f] = expandPath(conf.Applications[k].Files[f])
      }
    }

  }
	return conf
}

// Show prints the app configuration
func (c SetterConfig) Show() {
	fmt.Println("GithubToken: ", c.GithubToken)
	fmt.Println("SchemesListFile: ", c.SchemesListFile)
	fmt.Println("TemplatesListFile: ", c.TemplatesListFile)
	fmt.Println("SchemesCachePath: ", c.SchemesCachePath)
	fmt.Println("TemplatesCachePath: ", c.TemplatesCachePath)
	fmt.Println("DryRun: ", c.DryRun)

	for app, appConfig := range c.Applications {
		fmt.Println("  App: ", app)
		fmt.Println("    Enabled: ", appConfig.Enabled)
		fmt.Println("    Hook: ", appConfig.Hook)
		fmt.Println("    Comment_Prefix: ", appConfig.Comment_Prefix)
		for k, v := range appConfig.Files {
			fmt.Println("      ", k, "  ", v)
		}
	}
}
