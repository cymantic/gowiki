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
	flag.Parse()

	storage := NewFilePageStorage(*dataDir)
	renderer := NewRenderer("default")
	wiki := NewWiki(renderer, storage)

	port := ":" + strconv.Itoa(*ip)
	log.Info("starting wiki engine on localhost" + port)

	wiki.start(port)
}