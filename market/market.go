package market

type IMarket interface {
	SetPackage(packageID string)
	GetInfo() IPackage
}

type IPackage interface {
	List() []CodeExtension
	Versions(extID string) []string
	Latest() string
	Download(version string) (bool, error)
	Name() string
	Publisher() string
}
