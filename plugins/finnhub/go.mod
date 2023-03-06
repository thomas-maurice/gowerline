module github.com/thomas-maurice/gowerline/plugins/finnhub

go 1.18

replace github.com/thomas-maurice/gowerline/gowerline-server => ../../gowerline-server

require (
	github.com/Finnhub-Stock-API/finnhub-go v1.2.1
	github.com/thomas-maurice/gowerline/gowerline-server v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.21.0
)

require (
	github.com/antihax/optional v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/sys v0.5.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
