package domain

type apiConsts struct {
	DefaultFolderName   string
	SearchTypeDashboard string
	SearchTypeFolder    string
}

var ApiConsts = apiConsts{
	DefaultFolderName:   "General",
	SearchTypeDashboard: "dash-db",
	SearchTypeFolder:    "dash-folder",
}
