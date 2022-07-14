## Ravn Palynological Database

### Building
Building has been tested with Go 1.17, but will likely work with 1.16.
Earlier versions will not work as go:embed is used. Running `go build`
in the project directory should be all that is required.

### Running
Source data is included in the data directory for convenience. Required
bleve indexes are built from the source data via the following commands
executed from the project directory:

    ./ravn genera data/taxon\ genera\ files/*.xlsx
    ./ravn references data/taxon\ references/*.txt
    ./ravn species -images data/taxon\ pictures data/taxon\ species\ files/*.xlsx

Once the required indexes are built, you can start the application:

    ./ravn start -images "data/taxon pictures"

Additional options are available for all the above commands, use
`./ravn <command> -help` for more information.
