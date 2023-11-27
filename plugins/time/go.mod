module github.com/thomas-maurice/gowerline/plugins/time

go 1.18

replace github.com/thomas-maurice/gowerline/gowerline-server => ../../gowerline-server

require (
	github.com/thomas-maurice/gowerline/gowerline-server v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.26.0
)

require (
	go.etcd.io/bbolt v1.3.8 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
