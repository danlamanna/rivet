module github.com/danlamanna/rivet

go 1.14

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/burntsushi/toml v0.3.1
	github.com/hashicorp/go-retryablehttp v0.6.6 // v0.6.7 causes too many open files on test-dkc.sh
	github.com/hashicorp/go-version v1.2.1
	github.com/sirupsen/logrus v1.6.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)
