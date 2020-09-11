package gist

import (
	"fmt"
	"vscode-ext-download/request"

	"github.com/segmentio/encoding/json"
)

type Gist struct {
	gistID   string
	response GistResponse
}

type GistResponse struct {
	Files struct {
		CloudSettings GistFile `json:"cloudSettings"`
		Extensions    GistFile `json:"extensions.json"`
	} `json:"files"`
}

type GistFile struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

type Extension struct {
	Metadata struct {
		PublisherId          string `json:"publisherId"`
		PublisherDisplayName string `json:"publisherDisplayName"`
	} `json:"metadata"`

	Name      string `json:"name"`
	Publisher string `json:"publisher"`
	Version   string `json:"version"`
}

func (gist *Gist) Load() error {
	req := request.New()
	req.SetHeader("Accept", "application/vnd.github.v3+json")

	response, err := req.Get(fmt.Sprintf("https://api.github.com/gists/%s", gist.gistID))
	if err != nil {
		return err
	}

	var gistResponse GistResponse
	if err := json.Unmarshal(response.Data, &gistResponse); err != nil {
		return err
	}

	gist.response = gistResponse
	return nil
}

func (gist *Gist) Extensions() []Extension {
	var extensionContent []Extension

	if err := json.Unmarshal([]byte(gist.response.Files.Extensions.Content), &extensionContent); err != nil {
		panic(err)
	}

	return extensionContent
}

func New(gistID string) *Gist {
	return &Gist{gistID: gistID}
}
