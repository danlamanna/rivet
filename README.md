# rivet
a tiny cli for syncing a directory with a girder instance

[![asciicast](https://asciinema.org/a/263615.svg)](https://asciinema.org/a/263615)

# installation

## mac
```
sudo curl -sL -o /usr/local/bin/rivet https://github.com/danlamanna/rivet/releases/download/v0.0.4/rivet-0.0.4-darwin-amd64
sudo chmod +x /usr/local/bin/rivet
```

## linux
```
sudo curl -sL -o /usr/local/bin/rivet https://github.com/danlamanna/rivet/releases/download/v0.0.4/rivet-0.0.4-linux-amd64
sudo chmod +x /usr/local/bin/rivet
```

# usage
```
rivet sync --auth "username:password" --url data.kitware.com path/to/local/dir girder://somegirderfolderid
```

to avoid passing credentials multiple times, use `rivet configure`.

# limitations
Due to the difficulty in representing Girder items in the context of a POSIX filesystem, items 
with 0 files and items with multiple files are ignored. There is no way to use rivet to upload
or download these.

Additionally, rivet doesn't attempt to sync item or folder metadata. It's purely a tool for
syncing blobs of data to their respective folders.

If you require support for these use cases, consider a more comprehensive tool such as
[girder-client](https://pypi.org/project/girder-client/).
