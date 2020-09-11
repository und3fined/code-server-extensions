package market

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/thoas/go-funk"

	"vscode-ext-download/request"
)

const (
	marketAPI         = "https://marketplace.visualstudio.com/_apis/public/gallery/extensionquery"
	marketURL         = "https://marketplace.visualstudio.com/items"
	marketDownloadURL = "https://marketplace.visualstudio.com/_apis/public/gallery/publishers"
)

var req = request.New()

type CodePackageInfo struct {
	extensions []CodeExtension
}

type CodeMarket struct {
	pkgs []string
}

type CodeFilterCriteria struct {
	FilterType int    `json:"filterType"`
	Value      string `json:"value"`
}

type CodeRequestFilter struct {
	Criteria    []CodeFilterCriteria `json:"criteria"`
	Direction   int                  `json:"direction"`
	PageSize    int                  `json:"pageSize"`
	PageNumber  int                  `json:"pageNumber"`
	SortBy      int                  `json:"sortBy"`
	SortOrder   int                  `json:"sortOrder"`
	PagingToken *string              `json:"pagingToken"`
}

type CodeRequest struct {
	AssetTypes *string             `json:"assetTypes"`
	Filters    []CodeRequestFilter `json:"filters"`
	Flags      int                 `json:"flags"`
}

type CodePublisher struct {
	ID          string `json:"publisherId"`
	Name        string `json:"publisherName"`
	DisplayName string `json:"displayName"`
	Flags       string `json:"flags"`
}

type CodeVersion struct {
	Version     string    `json:"version"`
	Flags       string    `json:"flags"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type CodeExtension struct {
	Publisher     CodePublisher `json:"publisher"`
	ExtensionID   string        `json:"extensionId"`
	ExtensionName string        `json:"extensionName"`
	DisplayName   string        `json:"displayName"`
	Flags         string        `json:"flags"`
	LastUpdated   time.Time     `json:"lastUpdated"`
	PublishedDate time.Time     `json:"publishedDate"`
	ReleaseDate   time.Time     `json:"releaseDate"`
	Versions      []CodeVersion `json:"versions"`
}

type CodeResponse struct {
	Results []struct {
		Extensions []CodeExtension `json:"extensions"`
	} `json:"results"`
}

var _ IPackage = CodePackageInfo{}

func (market *CodeMarket) generateReferer(pkg string) string {
	return fmt.Sprintf("%s?itemName=%s", marketURL, pkg)
}

func (market *CodeMarket) SetPackage(pkgID string) {
	market.pkgs = append(market.pkgs, pkgID)
}

func (market *CodeMarket) GetInfo() IPackage {
	referer := market.generateReferer(market.pkgs[0])
	req.SetReferer(referer)

	var reqBody = CodeRequest{}
	reqBody.Flags = 103
	reqBody.Filters = append(reqBody.Filters, CodeRequestFilter{
		Direction:  2,
		PageSize:   100,
		PageNumber: 1,
		SortBy:     0,
		SortOrder:  0,
	})

	for _, pkg := range market.pkgs {
		reqBody.Filters[0].Criteria = append(reqBody.Filters[0].Criteria, CodeFilterCriteria{
			FilterType: 7,
			Value:      pkg,
		})
	}

	reqBodyBytes, err := json.Marshal(reqBody)

	if err != nil {
		panic(err)
	}

	req.SetHeader("Accept", "application/json;api-version=6.1-preview.1;excludeUrls=true")
	req.SetHeader("Accept-Language", "en-US,en;q=0.5")
	req.SetHeader("Content-Type", "application/json")

	req.SetBody(reqBodyBytes)

	response, err := req.Post(marketAPI)

	if err != nil {
		panic(err)
	}

	var codeResponse CodeResponse

	if err := json.Unmarshal(response.Data, &codeResponse); err != nil {
		panic(err)
	}

	extensions := codeResponse.Results[0].Extensions

	var packageExtensions CodePackageInfo

	for _, extension := range extensions {
		packageExtensions.extensions = append(packageExtensions.extensions, extension)
	}

	return packageExtensions
}

func (pkg CodePackageInfo) List() []CodeExtension {
	return pkg.extensions
}

func (pkg CodePackageInfo) Versions(extensionID string) []string {
	extension := funk.Find(pkg.extensions, func(ext CodeExtension) bool {
		return ext.ExtensionID == extensionID
	}).(CodeExtension)

	return funk.Map(extension.Versions, func(ver CodeVersion) string {
		return ver.Version
	}).([]string)
}

func (pkg CodePackageInfo) Latest() string {
	return ""
}

func (pkg CodePackageInfo) Download(version string) (bool, error) {
	return false, nil
}

func (pkg CodePackageInfo) Name() string {
	return ""
}

func (pkg CodePackageInfo) Publisher() string {
	return ""
}

func (ext *CodeExtension) Latest() string {
	return ext.Versions[0].Version
}

func (ext *CodeExtension) Download() error {
	extensionURL := fmt.Sprintf("/%s/vsextensions/%s/%s/vspackage", ext.Publisher.Name, ext.ExtensionName, ext.Latest())
	// req.SetHeader("X-VSS-ReauthenticationAction", "Suppress")
	response, err := req.Get(fmt.Sprintf("%s%s", marketDownloadURL, extensionURL))

	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(response.Filename, response.Data, 0644); err != nil {
		log.Println(err)
	}

	return nil
}

func NewCode() IMarket {
	return &CodeMarket{}
}
