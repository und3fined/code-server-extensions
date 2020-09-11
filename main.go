package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/theckman/yacspin"
	"github.com/thoas/go-funk"

	"vscode-ext-download/gist"
	"vscode-ext-download/market"
)

const gistID = "<Put your GIST ID here>"

func main() {
	// create spinner
	spinner, _ := yacspin.New(yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[52],
		Suffix:          " Process",
		SuffixAutoColon: true,
		Message:         "initial",
		StopCharacter:   "✓",
		StopColors:      []string{"fgGreen"},
	})

	spinner.Start()
	codeGist := gist.New(gistID)

	spinner.Message(fmt.Sprintf("Load gist %s", gistID))
	if err := codeGist.Load(); err != nil {
		panic(err)
	}

	spinner.Message(fmt.Sprintf("Load extensions from gist %s", gistID))
	gistExtensions := codeGist.Extensions()

	vscode := market.NewCode()
	totalExt := 0

	for i, ext := range gistExtensions {
		extName := fmt.Sprintf("%s.%s", ext.Publisher, ext.Name)
		vscode.SetPackage(extName)

		spinner.Message(fmt.Sprintf("Add extension %s to queue", extName))
		totalExt = (i + 1)
	}

	// // get info of extension
	spinner.Message(fmt.Sprintf("Get information of %d extensions", totalExt))
	pkgs := vscode.GetInfo()

	extensions := pkgs.List()
	var wg sync.WaitGroup

	spinner.Message(fmt.Sprintf("Get success %d extensions", len(extensions)))

	extensionUpdateList := []market.CodeExtension{}

	for _, ext := range extensions {
		gistExtRaw := funk.Find(gistExtensions, func(gistExt gist.Extension) bool {
			extName := strings.ToLower(ext.ExtensionName)
			extPublisher := strings.ToLower(ext.Publisher.Name)

			return strings.ToLower(gistExt.Name) == extName && strings.ToLower(gistExt.Publisher) == extPublisher
		})

		if gistExtRaw != nil {
			gistExt := gistExtRaw.(gist.Extension)
			if ext.Latest() != gistExt.Version {
				extensionUpdateList = append(extensionUpdateList, ext)
			}
		} else {
			log.Printf("Not found: %s", ext.DisplayName)
			log.Printf("Not ExtensionName: %s", ext.ExtensionName)
			log.Printf("Not Name: %s", ext.Publisher.Name)
		}

	}

	if len(extensionUpdateList) == 0 {
		spinner.Stop()
		log.Println("No extension update")
	} else {
		spinner.Message(fmt.Sprintf("Update %d extensions", len(extensionUpdateList)))

		for i := range extensionUpdateList {
			wg.Add(1)

			go func(ext *market.CodeExtension) {
				spinner.Message(fmt.Sprintf("Download %s", ext.DisplayName))
				if err := ext.Download(); err != nil {
					panic(err)
				}
				spinner.Message(fmt.Sprintf("Downloaded %s", ext.DisplayName))
				wg.Done()
			}(&extensions[i])
		}

		wg.Wait()
		spinner.Stop()
	}
}
