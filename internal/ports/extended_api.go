package ports

type ExtendedApi interface {
	GetConfiguredOrgId(orgName string) (int64, error)
}
