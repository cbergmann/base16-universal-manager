package main

import (
	"fmt"
	"github.com/hoisie/mustache"
	"gopkg.in/alecthomas/kingpin.v2"
  "go4.org/xdgdir"
	"os"
	"path/filepath"
)

// Configuration file
var configFile string

//Flags
var (
	updateFlag         = kingpin.Flag("update-list", "Update the list of templates and colorschemes").Bool()
	clearListFlag      = kingpin.Flag("clear-list", "Delete local master list caches").Bool()
	clearSchemesFlag   = kingpin.Flag("clear-templates", "Delete local scheme caches").Bool()
	clearTemplatesFlag = kingpin.Flag("clear-schemes", "Delete local template caches").Bool()
	configFileFlag     = kingpin.Flag("config", "Specify configuration file to use").Default(filepath.Join(xdgdir.Config.Path(),"base16-universal-manager/config.yaml")).String()
	schemeFlag         = kingpin.Flag("scheme", "Override scheme from config file").Default("").String()
)

//Configuration
var appConf SetterConfig

func main() {

	//Pase Flags
	kingpin.Version("1.0.0")
	kingpin.Parse()

	appConf = NewConfig(*configFileFlag)

	// appConf.Show()
	//TODO delete caches, if user wants to

	//Create cache paths, if missing
	os.MkdirAll(appConf.SchemesCachePath, os.ModePerm)
	os.MkdirAll(appConf.TemplatesCachePath, os.ModePerm)

	schemeList := LoadBase16ColorschemeList()
	templateList := LoadBase16TemplateList()

	if *updateFlag {
		schemeList.UpdateSchemes()
		templateList.UpdateTemplates()
	}

  schemename := appConf.Colorscheme

  if *schemeFlag != "" {
    schemename = *schemeFlag
  }

	scheme := schemeList.Find(schemename)
	fmt.Println("[CONFIG]: Selected scheme: ", scheme.Name)

	for k := range appConf.Applications {

		schemeList = LoadBase16ColorschemeList()
		templateList = LoadBase16TemplateList()

		templ := templateList.Find(k)

		Base16Render(templ, scheme)

	}

}

func Base16Render(templ Base16Template, scheme Base16Colorscheme) {

	fmt.Println("[RENDER]: Rendering template \"" + templ.Name + "\"")

  for k, v := range templ.Files {
		templFileData, err := DownloadFileToStirng(templ.RawBaseURL + "templates/" + k + ".mustache")
		check(err)
		configPath := appConf.Applications[templ.Name].Files[k]
    if configPath == "" {
			fmt.Println("     - skipping file because it is not configured: ", k)
      continue
    }

    renderedFile := mustache.Render(templFileData, scheme.MustacheContext())

    savePath := filepath.Join(configPath, k + v.Extension)

    if stat, err := os.Stat(configPath); err == nil && ! stat.IsDir() {
      //if the file exists and is a File write to it directly
      savePath = configPath
    }

    os.MkdirAll(filepath.Dir(savePath), os.ModePerm)

    //If DryRun is enabled, just print the output location for debugging
    if appConf.DryRun {
			fmt.Println("    - (dryrun) file would be written to: ", savePath)
		} else {
			switch appConf.Applications[templ.Name].Mode {
			case "rewrite":
				fmt.Println("     - writing: ", savePath)
				saveFile, err := os.Create(savePath)
				defer saveFile.Close()
				check(err)
				saveFile.Write([]byte(renderedFile))
				saveFile.Close()
			case "append":
				fmt.Println("     - appending to: ", savePath)
			case "replace":
				fmt.Println("     - replacing in: ", savePath)
			}
		}
	}

	if appConf.DryRun {
		fmt.Println("Not running hook, DryRun enabled: ", appConf.Applications[templ.Name].Hook)
	} else {
		exe_cmd(appConf.Applications[templ.Name].Hook)
	}
}

//TODO proper error handling
func check(e error) {
	if e != nil {
		panic(e)
	}

}
