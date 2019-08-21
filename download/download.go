package download

import (
	"fmt"
	"os"
	"path"

	"github.com/danlamanna/rivet/girder"

	sync_ "sync"
)

func maybeDownloadItem(ctx *girder.Context, p *girder.PathAndResource) {
	files := girder.ItemFiles(ctx, p.Resource.GirderID)

	if len(files) == 0 {
		ctx.Logger.Debugf("skipping sync of item, 0 files found")
		return
	} else if len(files) > 1 {
		ctx.Logger.Warnf("skipping sync of item, > 1 files found")
		return
	}
	err := os.MkdirAll(path.Dir(p.Path), os.ModePerm)
	if err != nil {
		ctx.Logger.Errorf("failed to create local directory %s, err: %s", path.Dir(p.Path), err)
	}
	st, err := os.Stat(p.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Errorf("failed to stat %s, err: %s", p.Path, err)
			return
		}
	}
	if st != nil && st.Size() == files[0].Size {
		ctx.Logger.Debugf("skipping (same size) %s\n", p.Path)
		return
	}
	out, err := os.Create(p.Path)
	if err != nil {
		ctx.Logger.Errorf("failed to create local file %s, err: %s", p.Path, err)
	}
	defer out.Close()
	ctx.Logger.Infof("downloading %s -> %s\n", p.Resource.GirderID, p.Path)
	_, err = girder.GetDownload(ctx, fmt.Sprintf("item/%s/download", p.Resource.GirderID), out)
	if err != nil {
		ctx.Logger.Errorf("failed to download file %s, err: %s", p.Path, err)
	}

}

func downloadFolder(ctx *girder.Context, src girder.GirderID, dest string, itemsToDownload chan *girder.PathAndResource) {

	// queue items for download
	for _, item := range girder.Items(ctx, src) {
		p := new(girder.PathAndResource)
		p.Path = path.Join(dest, item.Name)
		p.Resource = new(girder.Resource)
		p.Resource.GirderID = item.ID
		itemsToDownload <- p
	}

	// recurse on folders
	for _, folder := range girder.Folders(ctx, src) {
		// make folder (empty dir case)
		err := os.MkdirAll(path.Join(dest, folder.Name), os.ModePerm)
		if err != nil {
			ctx.Logger.Errorf("failed to create local directory %s, err: %s", path.Join(dest, folder.Name), err)
		}
		downloadFolder(ctx, folder.ID, path.Join(dest, folder.Name), itemsToDownload)
	}
}

func Download(ctx *girder.Context, src girder.GirderID, dest string) {
	itemsToDownload := make(chan *girder.PathAndResource)
	var wg sync_.WaitGroup

	for w := 1; w <= 10; w++ {
		go func() {
			for pathAndResource := range itemsToDownload {
				wg.Add(1)
				maybeDownloadItem(ctx, pathAndResource)
				wg.Done()
			}
		}()
	}

	downloadFolder(ctx, src, dest, itemsToDownload)

	wg.Wait()
	close(itemsToDownload)

	fmt.Println("done")
	return
}
