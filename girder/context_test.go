package girder

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestContext_ValidateURL(t *testing.T) {
	type fields struct {
		Auth        string
		URL         string
		Logger      *logrus.Logger
		Destination string
		ResourceMap ResourceMap
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"no scheme",
			fields{URL: "data.kitware.com/api/v1",
				Logger: logrus.New()},
		},
		{
			"no scheme, no url",
			fields{URL: "data.kitware.com",
				Logger: logrus.New()},
		},
		{
			"no scheme, no url, trailing slash",
			fields{URL: "data.kitware.com/",
				Logger: logrus.New()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Auth:        tt.fields.Auth,
				URL:         tt.fields.URL,
				Logger:      tt.fields.Logger,
				Destination: tt.fields.Destination,
				ResourceMap: tt.fields.ResourceMap,
			}
			c.ValidateURL()

			if c.URL != "https://data.kitware.com/api/v1" {
				t.Error(c.URL)
			}

		})
	}
}
