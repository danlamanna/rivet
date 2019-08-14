package templates

var DefaultUsageTemplate = `NAME
	rivet - a utility for syncing files to girder

SYNOPSIS
	rivet configure
	rivet sync source-directory girder://girder-folder-id
	rivet version

CONFIGURATION
	rivet needs to know about connecting to a remote girder instance. The
	--auth and --url flag can be used to convey this information to every
	subcommand. Alternatively, environment variables may be used, RIVET_AUTH
	and RIVET_URL respectively. If both of those options have been exhausted,
	rivet will use a configuration file. To create this configuration file
	you may run the rivet configure command and answer the prompted questions.

SUBCOMMANDS
	See rivet help configure and rivet help sync.

OPTIONS
	-a, --auth 
	    Credentials for authenticating with a remote girder instance. This may
	    be a username:password pair, an API key, or an existing token. Credentials
	    are immediately exchanged for a temporary token. This overrides the RIVET_AUTH
	    environment variable.

	-u, --url 
	    A location to a remote girder instance e.g. data.kitware.com, 
	    https://some-girder-instance.org/api/v1. The scheme is assumed
	    to be https if not otherwise passed. This overrides the RIVET_URL environment
	    variable.

	-v, --verbose 
	    Displays extra debugging information. If passed once it will set the log level
	    to debug, if set twice it will set it to trace. Trace is particularly noisy and 
	    will print every http request made by rivet.
`

var ConfigureUsageTemplate = `SYNOPSIS
	rivet configure

DESCRIPTION
	Configures a default profile to use for future calls to rivet. This
	saves authentication credentials and remote url of a girder instance
	to a configuration file at $HOME/.rivet/config.toml.

NOTES
	Environment variables such as RIVET_AUTH and RIVET_URL, as well as flags, will
	override settings configured with rivet configure.
`
var SyncUsageTemplate = `SYNOPSIS
	rivet sync source-directory girder://girder-folder-id

DESCRIPTION
	The sync command will copy files and folders from a local machine to a 
	folder on a remote girder instance. Sync uses the size of files to determine
	whether it can skip sending the data to the remote. Note that symbolic
	links will be skipped.

	The sync command requires a URL and set of credentials to use for connecting
	to the remote. These can be preconfigured with rivet configure, or passed 
	via --auth and --url. See rivet help for details.

USAGE
	For example:

		rivet sync ./local-folder girder://5d3bf0f6877dfcc902333a40

	This will sync all directories and files within local-folder to the girder
	folder with the ID of 5d3bf0f6877dfcc902333a40. That is, it will make the
	girder folder appear the same as local-folder.

	Running an identical command a second time should result in no changes, assuming
	the local and remote haven't been modified by any other tools.

NOTES
	Environment variables such as RIVET_AUTH and RIVET_URL, as well as flags, will
	override settings configured with rivet configure.
`
