# rivet
a tiny cli for syncing a directory with a girder instance

[![asciicast](https://asciinema.org/a/263615.svg)](https://asciinema.org/a/263615)

# installation

## mac
```
sudo curl -sL -o /usr/local/bin/rivet https://github.com/danlamanna/rivet/releases/download/v0.0.3/rivet-0.0.3-darwin-amd64
sudo chmod +x /usr/local/bin/rivet
```

## linux
```
sudo curl -sL -o /usr/local/bin/rivet https://github.com/danlamanna/rivet/releases/download/v0.0.3/rivet-0.0.3-linux-amd64
sudo chmod +x /usr/local/bin/rivet
```

# usage
```
rivet sync --auth "username:password" --url data.kitware.com path/to/local/dir girder://somegirderfolderid
```

to avoid passing credentials multiple times, use `rivet configure`.
