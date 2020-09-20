package runner

const (
  version = "1.4.0"
  author  = "dwisiswant0"
  banner  = `
   _____ _____ __    _____             
  |     | __  |  |  |   __|_ _ ___ ___ 
  |   --|    -|  |__|   __| | |- _|- _|
  |_____|__|__|_____|__|  |___|___|___|

      v` + version + ` - @` + author
  usage = `
  [buffers] | crlfuzz [options]
  crlfuzz [options]`
  options = `
  -u, --url <URL>           Define single URL to fuzz
  -l, --list <FILE>         Fuzz URLs within file
  -X, --method <METHOD>     Specify request method to use (default: GET)
  -o, --output <FILE>       File to save results
  -d, --data <DATA>         Define request data
  -H, --header <HEADER>     Pass custom header to target
  -x, --proxy <URL>         Use proxy to fuzz
  -c, --concurrent <i>      Set the concurrency level (default: 20)
  -s, --silent              Silent mode
  -v, --verbose             Verbose mode
  -V, --version             Show current CRLFuzz version
  -h, --help                Display its help
`
)
