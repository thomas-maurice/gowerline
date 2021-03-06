module github.com/thomas-maurice/gowerline/plugins/sample_plugin

go 1.18

replace github.com/thomas-maurice/gowerline/gowerline-server => ../../gowerline-server

require (
	github.com/thomas-maurice/gowerline/gowerline-server v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.21.0
)

require (
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/sys v0.0.0-20220429233432-b5fbb4746d32 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
