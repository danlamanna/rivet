package transfer

import (
	"strings"
	"testing"

	"github.com/danlamanna/rivet/girder"
	"github.com/sirupsen/logrus"
)

func TestUpload(t *testing.T) {
	dest := "girder://5d5f91e414f6f916735faffc"

	Upload(&girder.Context{
		Auth:        "admin:password",
		URL:         "http://localhost:8080/api/v1",
		Logger:      logrus.New(),
		Destination: strings.TrimPrefix(dest, "girder://"),
		ResourceMap: make(girder.ResourceMap),
	}, "etc/testdata/big_files", girder.GirderID(dest))
}
