module github.com/esnet/grafana-dashboard-manager

go 1.16

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/go-openapi/errors v0.20.0 // indirect
	github.com/go-openapi/strfmt v0.20.1 // indirect
	github.com/gosimple/slug v1.1.1
	github.com/grafana-tools/sdk v0.0.0-20210402150123-f7c763c3738c
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/thoas/go-funk v0.8.0
	golang.org/x/sys v0.0.0-20210403161142-5e06dd20ab57 // indirect
	gopkg.in/yaml.v2 v2.2.8

)

replace github.com/grafana-tools/sdk v0.0.0-20210402150123-f7c763c3738c => github.com/safaci2000/sdk v0.1.0
