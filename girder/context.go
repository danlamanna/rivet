package girder

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// Context stores the entire context needed to run a sync command
type Context struct {
	Auth        string
	URL         string
	Logger      *logrus.Logger
	Destination string
	ResourceMap ResourceMap
}

func GetValidURL(maybeInvalidURL string) (string, error) {
	tempCtx := new(Context)
	tempCtx.URL = maybeInvalidURL
	tempCtx.Logger = logrus.New()

	u, err := url.Parse(maybeInvalidURL)
	if err != nil {
		return "", err
	}

	// for URLs without schemes, just force https
	if u.Scheme == "" {
		u.Scheme = "https"
		tempCtx.URL = u.String()
		u, _ = url.Parse(tempCtx.URL)
	}

	release := new(GirderRelease)
	httpErr := new(GirderError)
	resp, err := Get(tempCtx, "describe", release, httpErr)

	if err != nil || resp.StatusCode != 200 {
		// try tacking /api/v1 on the end of the url
		u.Path = strings.TrimPrefix(u.Path, "/")
		u.Path = strings.TrimSuffix(u.Path, "/")
		u.Path += "/api/v1"
		tempCtx.URL = u.String()
		_, err := Get(tempCtx, "describe", release, httpErr)

		if err != nil {
			return "", fmt.Errorf("failed to connect to %s", maybeInvalidURL)
		}
	}

	return tempCtx.URL, nil
}

func (c *Context) CheckMinimumVersion() error {
	release := new(GirderRelease)
	httpErr := new(GirderError)
	resp, err := Get(c, "system/version", release, httpErr)

	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return httpErr
	}
	version := ""
	if release.Release != "" {
		version = release.Release
	} else if release.APIVersion != "" {
		version = release.APIVersion

	} else {
		return errors.New("unable to determine version of remote girder version")
	}
	parts := strings.SplitN(version, ".", 3)
	x, y := parts[0], parts[1]
	major, _ := strconv.Atoi(x)
	minor, _ := strconv.Atoi(y)
	if int(major) < 2 || (int(major) == 2 && minor < int(3)) {
		return fmt.Errorf("girder located at %s is version %s but rivet requires >= 2.3.0", c.URL, version)
	}
	return nil
}

func (c *Context) ValidateAuth() error {
	if strings.Index(c.Auth, ":") != -1 {
		// try username/password
		token := new(GirderTokenResponse)
		httpErr := new(GirderError)
		resp, err := GetBasicAuth(c, c.Auth, "user/authentication", token, httpErr)
		if err != nil {
			return err
		} else if resp.StatusCode != 200 {
			return httpErr
		}
		c.Auth = token.AuthToken.Token
		c.Logger.Debugf("authenticated with username/password (user %s)", token.User.Email)
		return nil
	} else if len(c.Auth) == 40 {
		// try api key
		token := new(GirderTokenResponse)
		httpErr := new(GirderError)
		resp, err := Post(c, fmt.Sprintf("api_key/token?key=%s", c.Auth), nil, token, httpErr)
		if err != nil {
			return err
		} else if resp.StatusCode != 200 {
			return httpErr
		}
		c.Auth = token.AuthToken.Token
		c.Logger.Debugf("authenticated with api key (user %s)", token.User.ID)
		return nil
	} else if len(c.Auth) == 64 {
		// try token
		user := new(GirderUser)
		httpErr := new(GirderError)
		resp, err := Get(c, "user/me", user, httpErr)
		if err != nil {
			return err
		} else if resp.StatusCode != 200 {
			return httpErr
		} else if user.Email == "" {
			// account for the "null user"
			return errors.New("failed to authenticate")

		}
		c.Logger.Debugf("authenticated as %s", user.Email)
		return nil
	} else {
		return errors.New("Unable to validate credentials, have they expired?")

	}

	return nil
}
