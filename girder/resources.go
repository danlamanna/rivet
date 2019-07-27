package girder

import (
	"errors"
	"fmt"
	"strings"
)

func GetOrCreateFolderRecursive(ctx *Context, path string) (GirderID, error) {
	parentID := GirderID(ctx.Destination)
	parts := strings.Split(path, "/")

	for i, part := range parts {
		partialPath := strings.Join(parts[0:i+1], "/")

		// check map if this dir has already been made (parents)
		if val, ok := ctx.ResourceMap[partialPath]; ok {
			if !val.SkipSync && val.GirderID != "" {
				parentID = val.GirderID
				continue
			} else if val.SkipSync {
				// skip this folder for the same reason its parent was skipped
				ctx.ResourceMap[path].GirderType = "folder"
				ctx.ResourceMap[path].SkipSync = true
				ctx.ResourceMap[path].SkipReason = ctx.ResourceMap[partialPath].SkipReason
				ctx.Logger.Warnf("skipping creation of %s since parent failed", path)
				return "", errors.New("parent")
			}
		}

		folder := new(GirderObject)
		httpErr := new(GirderError)
		url := fmt.Sprintf("folder?parentType=folder&reuseExisting=true&name=%s&parentId=%s", part, string(parentID))
		_, err := Post(ctx, url, nil, folder, httpErr)
		if err != nil {
			ctx.Logger.Errorf("problem creating %s, err: %s", partialPath, err)
			ctx.ResourceMap[path].GirderType = "folder"
			ctx.ResourceMap[path].SkipSync = true
			ctx.ResourceMap[path].SkipReason = err.Error()
			return "", err
		} else if httpErr.Message != "" {
			ctx.Logger.Errorf("problem creating %s, err: %s", partialPath, httpErr.Message)
			ctx.ResourceMap[path].GirderType = "folder"
			ctx.ResourceMap[path].SkipSync = true
			ctx.ResourceMap[path].SkipReason = httpErr.Error()
			return "", httpErr
		}
		parentID = folder.ID
		ctx.ResourceMap[path].GirderType = "folder"
		ctx.ResourceMap[path].GirderID = folder.ID
	}

	return parentID, nil
}

func GetOrCreateItem(ctx *Context, folderID GirderID, name string) (GirderID, error) {
	obj := new(GirderObject)
	httpErr := new(GirderError)
	_, err := Post(ctx, fmt.Sprintf("item?folderId=%s&name=%s&reuseExisting=true", folderID, name), nil, obj, httpErr)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to create item %s, err: %s", name, err))
	} else if httpErr.Message != "" {
		return "", errors.New(fmt.Sprintf("failed to create item %s, err: %s", name, httpErr.Message))
	}
	return obj.ID, nil
}

func ItemFiles(ctx *Context, itemID GirderID) []GirderFile {
	files := new([]GirderFile)
	httpErr := new(GirderError)
	_, err := Get(ctx, fmt.Sprintf("item/%s/files", itemID), files, httpErr)
	if err != nil {
		return nil
	} else if httpErr.Message != "" {
		fmt.Printf(httpErr.Message)
		return nil
	}
	return *files
}
