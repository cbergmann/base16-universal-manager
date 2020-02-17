package main

import (
	"fmt"
  "go4.org/xdgdir"
	"os"
	"bufio"
	"path/filepath"
	"strings"
	"github.com/hoisie/mustache"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Configuration file
var configFile string

//Flags
var (
	updateFlag            = kingpin.Flag("update-list", "Update the list of templates and colorschemes").Bool()
	clearListFlag         = kingpin.Flag("clear-list", "Delete local master list caches").Bool()
	listSchemesFlag       = kingpin.Flag("list-schemes", "List templates and exit").Bool()
	listTemplatesFlag     = kingpin.Flag("list-templates", "List schemes and exit").Bool()
	listTemplateFilesFlag = kingpin.Flag("list-templatefiles", "List files for template and exit").Default("").String()
	clearSchemesFlag      = kingpin.Flag("clear-schemes", "Delete local scheme caches").Bool()
	clearTemplatesFlag    = kingpin.Flag("clear-template", "Delete local template caches").Bool()
	configFileFlag        = kingpin.Flag("config", "Specify configuration file to use").Default(filepath.Join(xdgdir.Config.Path(),"base16-universal-manager/config.yaml")).String()
	schemeFlag            = kingpin.Flag("scheme", "Override scheme from config file").Default("").String()
  printConfigFlag    = kingpin.Flag("print-config", "Print current configuration").Bool()
)

//Configuration
var appConf SetterConfig

func main() {
	//Parse Flags
	kingpin.Version("0.2.1")
	kingpin.Parse()

	appConf = NewConfig(*configFileFlag)

	if *printConfigFlag {
		appConf.Show()
	}

	if *clearListFlag {
		err := os.Remove(appConf.SchemesListFile)
		if err == nil {
			fmt.Printf("Deleted cached colorscheme list %s\n", appConf.SchemesListFile)
		} else {
			fmt.Fprintf(os.Stderr, "Error deleting cached colorscheme list: %v\n", err)
		}
		err = os.Remove(appConf.TemplatesListFile)
		if err == nil {
			fmt.Printf("Deleted cached template list %s\n", appConf.TemplatesListFile)
		} else {
			fmt.Fprintf(os.Stderr, "Error deleting cached template list: %v\n", err)
		}
	}

	if *clearSchemesFlag {
		err := os.RemoveAll(appConf.SchemesCachePath)
		if err == nil {
			fmt.Printf("Deleted cached colorscheme list %s\n", appConf.SchemesCachePath)
		} else {
			fmt.Fprintf(os.Stderr, "Error deleting cached colorschemes: %v\n", err)
		}
	}

	if *clearTemplatesFlag {
		err := os.RemoveAll(appConf.TemplatesCachePath)
		if err == nil {
			fmt.Printf("Deleted cached templates %s\n", appConf.TemplatesCachePath)
		} else {
			fmt.Fprintf(os.Stderr, "Error deleting cached templates: %v\n", err)
		}
	}

	// Create cache paths, if missing
	os.MkdirAll(appConf.SchemesCachePath, os.ModePerm)
	os.MkdirAll(appConf.TemplatesCachePath, os.ModePerm)

  schemeList := LoadBase16ColorschemeList()
	templateList := LoadBase16TemplateList()

  if *updateFlag {
		schemeList.UpdateSchemes()
		templateList.UpdateTemplates()
	}

  listing := false

  if *listSchemesFlag {
    schemeList.Print()
    listing = true
  }

  if *listTemplatesFlag {
    templateList.Print()
    listing = true
  }

  if *listTemplateFilesFlag != "" {
		templ := templateList.Find(*listTemplateFilesFlag)
    fmt.Printf("Files for %v:\n",templ.Name)
    for fk := range templ.Files {
      fmt.Println("    - ",fk)
    }
    listing = true
  }

  if listing {
     //exit after listing
     os.Exit(0)
  }

  schemename := appConf.Colorscheme
  if *schemeFlag != "" {
    schemename = *schemeFlag
  }

	scheme := schemeList.Find(schemename)
	fmt.Println("[CONFIG]: Selected scheme: ", scheme.Name)

	templateEnabled := false
	for app, appConfig := range appConf.Applications {
		if appConfig.Enabled {
			Base16Render(templateList.Find(app), scheme)
			templateEnabled = true
		}
	}

	if !templateEnabled {
		fmt.Println("No templates enabled")
	}

}

// Base16Render takes an application-specific template and renders a config file
// implementing the provided colorscheme.
func Base16Render(templ Base16Template, scheme Base16Colorscheme) {
  fmt.Println("[RENDER]: Rendering template \"" + templ.Name + "\"")

  for k, v := range templ.Files {
    var before, after string
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

    mode := appConf.Applications[templ.Name].Mode

    //If DryRun is enabled, just print the output location for debugging
    if appConf.DryRun {
			fmt.Println("    - (dryrun) file would be written to: ", savePath)
		} else {
			fmt.Fprintf(os.Stdout, "     - %ving %v in: %v\n", mode, k, savePath)

      start_line := tagline(templ, k, "start")
      end_line := tagline(templ, k, "end")

      //read file for append and replace mode
      write_start, write_end := false, false
			if mode == "append" || mode == "replace" {
        write_start = true

        file, err := os.Open(savePath)
        if err != nil {
		      fmt.Fprintf(os.Stderr, "Could not read file for %v mode: %v\n",mode, err)
          os.Exit(1)
        }

        state := "before"
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
          switch state {
          case "before":
            if scanner.Text() == start_line {
              state = "between"
            }else{
              before += scanner.Text() + "\n"
            }
          case "between":
            if scanner.Text() == end_line {
              state = "after"
              after = ""
            } else {
              after += scanner.Text() + "\n"
            }
          case "after":
            after += scanner.Text() + "\n"
          }
        }
        file.Close()

        //consistency check
        switch mode {
        case "replace":
          if state == "before" {
            fmt.Fprintf(os.Stderr, "For replace mode you have to add the following as a single line in target file:\n%v", start_line)
            os.Exit(1)
          } else if after != "" {
            write_end = true
          }
        case "append":
          //discard everything after start tag
          after = ""
        }
			}

      //write file out
      saveFile, err := os.Create(savePath)
			defer saveFile.Close()
			check(err)
			saveFile.Write([]byte(before))
			if write_start { saveFile.Write([]byte(start_line + "\n")) }
			saveFile.Write([]byte(renderedFile))
			if write_end { saveFile.Write([]byte(end_line + "\n")) }
			saveFile.Write([]byte(after))
			saveFile.Close()

		}
	}

	if appConf.DryRun {
		fmt.Println("Not running hook, DryRun enabled: ", appConf.Applications[templ.Name].Hook)
	} else {
    //allow string replacement
    replacer := strings.NewReplacer(
      "\\{", "{", //mark escaped curly braces
      "{template}", templ.Name,
      "{scheme}", scheme.Name,
    )
		exe_cmd(replacer.Replace(appConf.Applications[templ.Name].Hook))
    for k:= range(appConf.Applications[templ.Name].Hooks) {
      exe_cmd(replacer.Replace(appConf.Applications[templ.Name].Hooks[k]))
    }
	}
}

//TODO proper error handling
func check(e error) {
	if e != nil {
		panic(e)
	}

}

func tagline(templ Base16Template, file string, part string) string{
  return appConf.Applications[templ.Name].Comment_Prefix + templ.Name + "-" + file + "-" + part
}
