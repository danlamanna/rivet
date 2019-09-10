package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/danlamanna/rivet/config"
	"github.com/danlamanna/rivet/girder"
	"github.com/danlamanna/rivet/transfer"
	"github.com/danlamanna/rivet/version"
	log "github.com/sirupsen/logrus"
)

func Configure(ctx *girder.Context) {
	reader := bufio.NewReader(os.Stdin)
	var promptedURL string
	for {
		fmt.Print("girder url (e.g. data.kitware.com): ")
		promptedURL, _ = reader.ReadString('\n')
		promptedURL = strings.TrimSpace(promptedURL)

		if promptedURL != "" {
			break
		}
	}

	validURL, err := girder.GetValidURL(promptedURL)

	if err != nil {
		ctx.Logger.Fatal(err)
	}

	ctx.URL = validURL

	if err = ctx.CheckMinimumVersion(); err != nil {
		log.Fatal(err)
	}
	var promptedAuth string
	for {
		fmt.Print("auth credentials (e.g. username:password, token, api-key): ")
		promptedAuth, _ = reader.ReadString('\n')
		promptedAuth = strings.TrimSpace(promptedAuth)

		if promptedAuth != "" {
			break
		}
	}
	ctx.Auth = promptedAuth
	if err = ctx.ValidateAuth(); err != nil {
		log.Fatal(err)
	}
	config.WriteDefaultProfile(promptedAuth, validURL)
}

func Sync(ctx *girder.Context, source *string, dest *string) {
	*source = strings.TrimSuffix(*source, "/")
	*dest = strings.TrimSuffix(*dest, "/")

	sourceIsGirder := strings.HasPrefix(*source, "girder://")
	destIsGirder := strings.HasPrefix(*dest, "girder://")
	if sourceIsGirder && destIsGirder {
		log.Fatal("cannot sync between two girder instances")
	} else if !sourceIsGirder && !destIsGirder {
		log.Fatal("cannot sync between two local directories")
	}
	if destIsGirder {
		if stat, err := os.Stat(*source); err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("source directory %s does not exist.\n", *source)
			} else {
				log.Fatalf("failed to access source directory %s, err: %s.\n", *source, err)
			}
		} else if !stat.IsDir() {
			log.Fatalf("source %s is not a directory.\n", *source)
		}
	}

	ctx.ResourceMap = make(girder.ResourceMap)
	ctx.Destination = strings.TrimPrefix(*dest, "girder://")

	if err := ctx.CheckMinimumVersion(); err != nil {
		log.Fatal(err)
	}
	if err := ctx.ValidateAuth(); err != nil {
		log.Fatal(err)
	}

	if destIsGirder {
		transfer.Upload(ctx, *source, girder.GirderID(*dest))
	} else if sourceIsGirder {
		transfer.Download(ctx, girder.GirderID(strings.TrimPrefix(*source, "girder://")), *dest)
	}
}

func Version() {
	fmt.Printf("rivet       v%s\n", version.Version)
	fmt.Printf("build:      %s\n", version.GitCommit)
	fmt.Printf("built:      %s\n", version.BuildDate)
	fmt.Printf("go version: %s\n", version.GoVersion)
	fmt.Printf("os/arch:    %s\n", version.OsArch)
}

func APICreateFolder(ctx *girder.Context, dest string, path string) {
	if err := ctx.ValidateAuth(); err != nil {
		log.Fatal(err)
	}
	ctx.ResourceMap = make(girder.ResourceMap)
	ctx.ResourceMap[path] = new(girder.Resource)
	ctx.Destination = dest
	id, _ := girder.GetOrCreateFolderRecursive(ctx, path)
	fmt.Println(id)
}
