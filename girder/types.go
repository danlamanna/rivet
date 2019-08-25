package girder

import (
	"path"
	"sync"
)

type GirderID string

type GirderObject struct {
	ID   GirderID `json:"_id"`
	Name string   `json:"name"`
}

type GirderTokenResponse struct {
	AuthToken struct {
		Token string `json:"token"`
	} `json:"authToken"`
	User struct {
		ID    string `json:"_id"`
		Email string `json:"email"`
	} `json:"user"`
}

type GirderUser struct {
	Email string `json:"email"`
}

type GirderRelease struct {
	Release    string `json:"release"`
	APIVersion string `json:"apiVersion"`
}

type FailureSummary map[string]string

// ResourceMap maps on-disk paths to "resources"
type ResourceMap map[string]*Resource

// Resource represents a local path and the relationship to a Girder resource
type Resource struct {
	Path string
	Size int64
	Type string
	Children []*Resource
	
	GirderParentID GirderID
	GirderID   GirderID
	GirderType string

	SkipSync   bool
	SkipReason string
}

func (m ResourceMap) Parent(resource *Resource) *Resource {
	return m[path.Dir(resource.Path)]
}

type GirderFile struct {
	ID   GirderID `json:"_id"`
	Size int64    `json:"size"`
}

type GirderError struct {
	Message string `json:"message"`
}

func (error *GirderError) Error() string {
	return error.Message
}

type ItemMap struct {
	sync.Mutex
	M map[string]GirderID
}

type PathAndResource struct {
	Path     string
	Resource *Resource
}
