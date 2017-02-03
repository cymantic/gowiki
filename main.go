package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

func init() {
	Environment := os.Getenv("ENV")
	if strings.ToLower(Environment) == "prod" {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.WarnLevel)
	} else {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stdout)
	}
}

func main() {
	var ip = flag.Int("port", 8080, "Web server port.")
	var dataDir = flag.String("data", "data", "Wiki Data Directory")
	var tmplDir = flag.String("tmpl", "tmpl", "Wiki Templtes Directory")
	var cloneFromGitRepo = flag.String("clone", "", "Clone from repository")
	var initFromGitRepo = flag.String("init", "", "Initialise from repository")
	var originGitRepo = flag.String("origin", "", "Initialise to repository")
	flag.Parse()

	wikiRepository, err := NewFileWikiRepository(*dataDir, *cloneFromGitRepo, *initFromGitRepo, *originGitRepo)
	if err != nil {
		log.Fatal("Can't initialise wiki data - ", err)
		return
	}

	templateRenderer := NewTemplateRenderer(*tmplDir, "default")
	wiki := NewWiki(wikiRepository, templateRenderer)

	port := ":" + strconv.Itoa(*ip)
	log.Info("starting wiki engine on localhost" + port + " from directory " + *dataDir + ".")

	err = wiki.start(port)
	if err != nil {
		log.Fatal(err)
	}
}
