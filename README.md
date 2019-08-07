# rivet
a tiny cli for syncing a directory with a girder instance

# usage
```
rivet sync --auth "username:password" --url data.kitware.com path/to/local/dir girder://somegirderfolderid
```

to avoid passing credentials multiple times, use `rivet configure`.
