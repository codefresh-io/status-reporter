module github.com/codefresh-io/status-reporter

go 1.15

require (
	cloud.google.com/go v0.74.0 // indirect
	github.com/argoproj/argo v0.0.0-20201214002321-380268943efc
	github.com/emicklei/go-restful v2.15.0+incompatible // indirect
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.3.0
	github.com/go-openapi/spec v0.20.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.3 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mitchellh/mapstructure v1.4.0 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9 // indirect
	golang.org/x/sys v0.0.0-20201211090839-8ad439b19e0f // indirect
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.8
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd // indirect
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.17.8

replace k8s.io/apimachinery => k8s.io/apimachinery v0.17.8

replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0

replace sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.2.9
