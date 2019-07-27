package transfer

import (
	"strings"
	"testing"

	"github.com/danlamanna/rivet/girder"
	"github.com/sirupsen/logrus"
)

// func TestMax(t *testing.T) {
// 	if max(5, 6) != 6 {
// 		t.Error("dun work")
// 	}
// }

func BenchmarkUpload(b *testing.B) {
	dest := "girder://5d3bf0f6877dfcc902333a40"
	for n := 0; n < b.N; n++ {

		Upload(&girder.Context{
			Auth:        "thhUL4H6dkcBQuVsz7n5vVVjrDr3RJgiw8A4CaGjEVxhEO0ozjG7FVYld34tpm3Y",
			URL:         "https://data.kitware.com/api/v1",
			Logger:      logrus.New(),
			Destination: strings.TrimPrefix(dest, "girder://"),
			ResourceMap: make(girder.ResourceMap),
		}, "/Users/dan/p/rivet/.git", girder.GirderID(dest))
	}
}
