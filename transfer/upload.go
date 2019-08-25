package transfer

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	//"strings"
	//"sync"
    //"http"
	"github.com/danlamanna/rivet/girder"
	"github.com/danlamanna/rivet/util"
)

const maxChunkSize = 1024 * 1024 * 16

// build these synchronously or use a better data structure for determining when parents are created
func buildGirderDirs(ctx *girder.Context, baseDir string) {

	// get directories to build in sorted order (to avoid extraneous girder POST requests)
	dirsToBuild := make([]string, 0)
	for k, v := range ctx.ResourceMap {
		if v.Type == "directory" {
			dirsToBuild = append(dirsToBuild, k)
		}

	}
	sort.Slice(dirsToBuild, func(i, j int) bool { return dirsToBuild[i] < dirsToBuild[j] })

	for _, v := range dirsToBuild {
		girder.GetOrCreateFolderRecursive(ctx, v)
	}
}

// first round get the roots and do those
// after that can get each set of children from parents

func GetOrCreateFolder(ctx *girder.Context, folderResource *girder.Resource) (girder.GirderID, error) {
	//ctx.Logger.Debugf("GOCF %s", folderResource)
	//fmt.Println("GOCF", folderResource)
	//fmt.Println("parentID", string(folderResource.GirderParentID))
	//xxparentID := girder.GirderID(ctx.Destination)
	
	girderFolder := new(girder.GirderObject)
	/*
	folderName := path.Base(folderResource.Path)
    httpErr := new(girder.GirderError)
    url := fmt.Sprintf("folder?parentType=folder&reuseExisting=true&name=%s&parentId=%s", url.QueryEscape(folderName), string(folderResource.GirderParentID))
    _, err := girder.Post(ctx, url, nil, girderFolder, httpErr)
	if err != nil {
		ctx.Logger.Errorf("problem creating %s, err: %s", folderName, err)
		folderResource.GirderType = "folder"
		folderResource.SkipSync = true
		folderResource.SkipReason = err.Error()
		return "", err
	} else if httpErr.Message != "" {
		ctx.Logger.Errorf("problem creating %s, err: %s", folderName, httpErr.Message)
		folderResource.GirderType = "folder"
		folderResource.SkipSync = true
		folderResource.SkipReason = httpErr.Error()
		return "", httpErr
	}
	folderResource.GirderID = girderFolder.ID
	*/
	// mocking calls to girder server
	path.Base(folderResource.Path)
	folderResource.GirderType = "folder"
	folderResource.GirderID = "mock"
	// mock




	//fmt.Println("GOCF after", folderResource)
	return girderFolder.ID, nil
}

func buildGirderDirsP(ctx *girder.Context) {
    //fmt.Println("parallel")

	rootDirsToBuild := make([]*girder.Resource, 0)
	var numRootDirs, numDirs, numFiles int
	for _, v := range ctx.ResourceMap {
		if v.Type == "file" {
            numFiles++
		} else if v.Type == "directory" {
			numDirs++
    		parent := ctx.ResourceMap.Parent(v)
	    	if parent == nil {
				rootDirsToBuild = append(rootDirsToBuild, v)
				numRootDirs++
			}
		}
	}
    fmt.Println("about to process root nodes and total dirs", numRootDirs, numDirs)

	resourceJobs := make(chan *girder.Resource, numDirs)
	resourceResults := make(chan bool, numDirs)
    fmt.Println("length, capacity of resourceJobs", len(resourceJobs), cap(resourceJobs))
	 
	numWorkers := 10
    for w := 0; w < numWorkers; w++ {
		go func (id int, jobs chan *girder.Resource, results chan<- bool) {
			for resourceJob := range jobs {
				//fmt.Println(id, "postgrab length, capacity of resourceJobs", len(jobs), cap(jobs))
				if resourceJob != nil {
					//fmt.Printf("worker id %d", id)
					//fmt.Println(resourceJob)
					GetOrCreateFolder(ctx, resourceJob)
					// now add all children of this folder to the queue for concurrent processing
					fmt.Println("will add num children", len(resourceJob.Children))
					for _, v := range resourceJob.Children {
						currentChild := v
						currentChild.GirderParentID = girder.GirderID(resourceJob.GirderID)
						//fmt.Println(id, "child push", v.Path)
						go func () { 
							fmt.Println(id, "prepush length, capacity of resourceJobs", len(jobs), cap(jobs), currentChild.Path)
							jobs<-currentChild
						}()
						//jobs<-currentChild
						//fmt.Println(id, "child push returned", v.Path)
					}
					results <- true
				}
			}
		}(w, resourceJobs, resourceResults)	
    }

	fmt.Println("root add", len(rootDirsToBuild))
	for _, v := range rootDirsToBuild {
        v.GirderParentID = girder.GirderID(ctx.Destination)
		//fmt.Println("Adding to queue:", v)
		resourceJobs<-v
	}
	
	for a := 0; a < numDirs; a++ {
		<-resourceResults
		if a % 50 == 0 { fmt.Println("Got results ", a) }
    }

}

func _uploadBytes(ctx *girder.Context, upload girder.GirderID, fullPath string, fi os.FileInfo) {

	file, err := os.Open(fullPath)
	if err != nil {
		ctx.Logger.Warnf("failed to access %s, skipping. err: %s", fullPath, err)
		return
	}
	defer file.Close()

	totalChunks := util.Max(0, fi.Size()/maxChunkSize) + 1
	var offset int64
	i := 1
	for {
		bufSize := util.Min(maxChunkSize, fi.Size()-offset)
		buffer := make([]byte, bufSize)
		_, err := file.Read(buffer)

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}

			break
		}

		r := bytes.NewReader(buffer)
		chunkOrFile := new(girder.GirderObject)
		gerr := new(girder.GirderError)

		if totalChunks > 1 {
			ctx.Logger.Debugf("%s - uploading chunk %d/%d", fullPath, i, totalChunks)
		}
		_, err = girder.Post(ctx, fmt.Sprintf("file/chunk?uploadId=%s&offset=%d", upload, offset), r, chunkOrFile, gerr)

		offset, err = file.Seek(0, os.SEEK_CUR)
		if err != nil {
			// handle error
		}

		i++
		if offset >= fi.Size() {
			break
		}
	}
}

func uploadFile(ctx *girder.Context, parentID girder.GirderID, fullPath string, name string) int {
	files := girder.ItemFiles(ctx, parentID)
	upload := new(girder.GirderObject)
	gerr := new(girder.GirderError)

	fi, err := os.Stat(fullPath)
	if err != nil {
		ctx.Logger.Warnf("couldn't stat %s, skipping. err: %s", fullPath, err)
		return 0
	}
	if len(files) == 0 {
		ctx.Logger.Debugf("detected new file %s\n", fullPath)
		ctx.Logger.Infof("uploading: %s\n", fullPath)
		// creating a new file
		girder.Post(ctx, fmt.Sprintf("file?parentId=%s&name=%s&parentType=item&size=%d", parentID, url.QueryEscape(name), fi.Size()), nil, upload, nil)

		_uploadBytes(ctx, upload.ID, fullPath, fi)
	} else if len(files) == 1 {
		// potentially updating the contents of an existing file, or no-oping

		if files[0].Size != fi.Size() {
			ctx.Logger.Debugf("file sizes differ for %s\n", fullPath)
			ctx.Logger.Infof("uploading: %s\n", fullPath)
			// change file contents
			girder.Put(ctx, fmt.Sprintf("file/%s/contents?size=%d",

				files[0].ID, fi.Size()), nil, upload, gerr)

			_uploadBytes(ctx, upload.ID, fullPath, fi)
		}

	} else {
		fmt.Println("item has > 1 file.. not doing anything")
	}

	return 1
}

func buildResourceMap(ctx *girder.Context, baseDir string) (int, int) {
	numDirs, numFiles := 0, 0

	err := filepath.Walk(baseDir, func(filepath string, info os.FileInfo, err error) error {
        //ctx.Logger.Warnf("%s", filepath)
		if err != nil {
			ctx.Logger.Warnf("failed to access %s, skipping", filepath)
			return nil
		}

		if filepath == "." {
			return nil
		}

		fileType := ""
		if info.IsDir() {
			fileType = "directory"
			numDirs++
			// } else if s, _ := os.Readlink(filepath); s != nil {
			// 	ctx.Logger.Warnf("found symlink %s, skipping", filepath)
		} else {
			fileType = "file"
			numFiles++
		}

		currentNode := &girder.Resource{
			Path: filepath,
			Size: info.Size(),
			Type: fileType,
		}
		ctx.ResourceMap[filepath] = currentNode
        parent := ctx.ResourceMap.Parent(currentNode)
		if parent != nil {
			parent.Children = append(parent.Children, currentNode)
		}

		return nil
	})

	if err != nil {
		ctx.Logger.Errorf("failed to walk the %s directory", baseDir)
		os.Exit(1)
	}

	return numDirs, numFiles
}

func Upload(ctx *girder.Context, source string, destination girder.GirderID) {
	ctx.Logger.Debugf("scanning %s for syncable items", source)

	absSource, _ := filepath.Abs(source)
	os.Chdir(absSource)
	source = "."
	numDirs, numFiles := buildResourceMap(ctx, source)
	ctx.Logger.Infof("found %d dirs, %d files to potentially sync", numDirs, numFiles)

	destFolder := new(girder.GirderObject)
	httpErr := new(girder.GirderError)
	resp, err := girder.Get(ctx, fmt.Sprintf("folder/%s", ctx.Destination), destFolder, httpErr)

	if err != nil {
		ctx.Logger.Fatalf("failed to retrieve destination folder, err: %s", err)
	} else if resp.StatusCode != 200 {
		ctx.Logger.Fatalf("failed to retrieve destination folder, err: %s", httpErr.Message)
	}

	ctx.Logger.Info("building remote girder directories")
	
	//buildGirderDirs(ctx, source)
	buildGirderDirsP(ctx)

	ctx.Logger.Info("done building remote girder directories")

/*
	ctx.Logger.Info("building remote girder items")

	numJobs := 0
	var mutex sync.Mutex
	itemsToUpload := make(chan *girder.PathAndResource, numFiles)
	results := make(chan bool, numFiles)

	// spawn 10 workers for building items
	for w := 1; w <= 10; w++ {
		go func() {
			for pathAndResource := range itemsToUpload {
				parent := ctx.ResourceMap.Parent(pathAndResource.Resource)

				// default to the root sync dest, override if there's a parent
				parentID := girder.GirderID(strings.TrimPrefix(string(destination), "girder://"))
				if parent != nil {
					parentID = parent.GirderID
				}
				if parent != nil && parent.SkipSync {
					mutex.Lock()
					ctx.ResourceMap[pathAndResource.Path].SkipSync = true
					ctx.ResourceMap[pathAndResource.Path].SkipReason = parent.SkipReason
					mutex.Unlock()
					ctx.Logger.Warnf("skipping sync of %s because parent failed to be created", parent.Path)
					results <- true
					continue
				}
				itemID, err := girder.GetOrCreateItem(ctx, parentID, filepath.Base(pathAndResource.Path))
				mutex.Lock()
				if err != nil {
					ctx.ResourceMap[pathAndResource.Path].SkipSync = true
					ctx.ResourceMap[pathAndResource.Path].SkipReason = err.Error()
					ctx.Logger.Error(err)
				} else {
					ctx.ResourceMap[pathAndResource.Path].GirderID = itemID
				}
				mutex.Unlock()
				results <- true
			}
		}()
	}

	for filepath, resource := range ctx.ResourceMap {
		if resource.Type == "file" {
			f := new(girder.PathAndResource)
			f.Path = filepath
			f.Resource = resource
			numJobs++
			itemsToUpload <- f
		}
	}
	close(itemsToUpload)

	for a := 0; a < numJobs; a++ {
		<-results
	}

	ctx.Logger.Info("syncing blobs")

	numJobs = 0
	filesToUpload := make(chan *girder.PathAndResource, numFiles)
	results = make(chan bool, numFiles)

	// spawn 10 workers for uploading files
	for w := 1; w <= 10; w++ {
		go func() {
			for pathAndResource := range filesToUpload {
				if pathAndResource != nil {
					uploadFile(ctx, pathAndResource.Resource.GirderID, pathAndResource.Path, path.Base(pathAndResource.Path))
				}
				results <- true
			}
		}()
	}

	for filepath, resource := range ctx.ResourceMap {
		if resource.Type == "file" && resource.GirderID != "" {
			f := new(girder.PathAndResource)
			f.Path = filepath
			f.Resource = resource
			numJobs++
			filesToUpload <- f
		} else if resource.Type == "file" && resource.GirderID == "" {
			// it was printed as an error above
			ctx.Logger.Infof("skipping sync of %s because parent item creation failed.", filepath)
			numJobs++
			filesToUpload <- nil
		}
	}
	close(filesToUpload)

	for a := 0; a < numJobs; a++ {
		<-results
	}

	ctx.Logger.Info("")

	ctx.Logger.Info("summary:")

	var numSucceeded int
	var numFailed int

	failureSummary := make([]string, 0)
	for k, v := range ctx.ResourceMap {
		if v.SkipSync {
			numFailed++
			failureSummary = append(failureSummary, fmt.Sprintf("%s: %s", k, v.SkipReason))
		} else {
			numSucceeded++
		}
	}

	sort.Slice(failureSummary, func(i, j int) bool { return failureSummary[i] < failureSummary[j] })

	ctx.Logger.Infof("successfully synced %d files/folders", numSucceeded)

	if numFailed > 0 {
		ctx.Logger.Infof("failed to sync %d files/folders:", numFailed)

		for _, failure := range failureSummary {
			ctx.Logger.Info(failure)
		}
	}
*/
	return
}
