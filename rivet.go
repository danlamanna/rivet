package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/danlamanna/rivet/config"
	"github.com/danlamanna/rivet/girder"
	"github.com/danlamanna/rivet/transfer"
)

var version = "undefined"

var auth string
var url string
var verbose bool

func main() {

	defaultAuth := ""
	defaultURL := ""
	profile, err := config.ReadDefaultProfile()
	if profile != nil && err == nil {
		defaultAuth = profile.Auth
		defaultURL = profile.URL

	}

	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, V",
		Usage: "print only the version",
	}

	app := cli.NewApp()
	app.Version = version
	app.Name = "rivet"
	app.Usage = "sync files to girder"
	app.UsageText = "\033[1mrivet sync\033[0m [\033[1m-auv\033[0m] \033[4msource-directory\033[0m \033[4mgirder://girder-folder-id\033[0m"
	app.Description = `
	In the USAGE example, rivet will copy the contents of source-directory 
	to girder-folder-id, skipping files which are the same size on the
	remote.

	Syncing requires a set of credentials and the URL of the remote Girder
	instance. See the --auth and --url flag descriptions. 

	Files on the local filesystem have an item and file created for them on
	the remote.

	Symbolic links are unsupported and will be skipped. `
	app.Action = func(c *cli.Context) error {
		cli.ShowAppHelpAndExit(c, 0)
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "auth, a",
			Value: defaultAuth,
			Usage: `Authentication credentials, can be username:password, 
			 a token, or an api key`,
			Destination: &auth,
			EnvVar:      "RIVET_AUTH",
		},
		cli.StringFlag{
			Name:  "url, u",
			Value: defaultURL,
			Usage: `URL of the girder instance, e.g. data.kitware.com, 
			somedomain.com/api/v1`,
			Destination: &url,
			EnvVar:      "RIVET_URL",
		},
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "Add debug logging",
			Destination: &verbose,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "configure",
			Usage: "configure credentials for remote girder instances",
			Action: func(c *cli.Context) error {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("girder url (e.g. data.kitware.com): ")
				url, _ := reader.ReadString('\n')
				url = strings.TrimSpace(url)
				validURL, err := girder.GetValidURL(url)

				if err != nil {
					log.Fatal(err)
				}
				girderCtx := new(girder.Context)
				girderCtx.URL = validURL
				girderCtx.Logger = logrus.New()
				if err = girderCtx.CheckMinimumVersion(); err != nil {
					log.Fatal(err)
				}
				fmt.Print("auth credentials (e.g. username:password, token, api-key): ")
				auth, _ := reader.ReadString('\n')
				auth = strings.TrimSpace(auth)
				girderCtx.Auth = auth
				if err = girderCtx.ValidateAuth(); err != nil {
					log.Fatal(err)
				}
				config.WriteDefaultProfile(auth, validURL)
				return nil
			},
		},
		{
			Name:  "sync",
			Usage: "sync a local directory to a girder folder",
			Action: func(c *cli.Context) error {
				source, dest := c.Args().Get(0), c.Args().Get(1)
				source = strings.TrimSuffix(source, "/")

				if _, err := os.Stat(source); err != nil {
					if os.IsNotExist(err) {
						log.Fatalf("source directory %s does not exist.\n", source)
					} else {
						log.Fatalf("failed to access source directory %s, err: %s.\n", source, err)
					}
				}

				if source == "" && dest == "" || !strings.HasPrefix(dest, "girder://") {
					cli.ShowCommandHelpAndExit(c, "sync", 1)
				} else if auth == "" {
					fmt.Println("See --auth flag")
					os.Exit(1)
				} else if url == "" {
					fmt.Println("See --url flag")
					os.Exit(1)
				}

				validURL, err := girder.GetValidURL(url)

				if err != nil {
					log.Fatal(err)
				}

				girderCtx := new(girder.Context)
				girderCtx.Auth = auth
				girderCtx.URL = validURL
				girderCtx.Logger = logrus.New()
				girderCtx.ResourceMap = make(girder.ResourceMap)
				girderCtx.Destination = strings.TrimPrefix(dest, "girder://")
				if verbose {
					girderCtx.Logger.Level = logrus.DebugLevel
				} else {
					girderCtx.Logger.Level = logrus.InfoLevel
				}
				if err = girderCtx.CheckMinimumVersion(); err != nil {
					log.Fatal(err)
				}
				if err = girderCtx.ValidateAuth(); err != nil {
					log.Fatal(err)
				}

				transfer.Upload(girderCtx, source, girder.GirderID(dest))

				return nil
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
