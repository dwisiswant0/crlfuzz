package runner

import "github.com/projectdiscovery/gologger"

func showBanner() {
	gologger.Printf("%s\n\n", banner)
	gologger.Labelf("Use with caution. You are responsible for your actions\n")
	gologger.Labelf("Developers assume no liability and are not responsible for any misuse or damage.\n")
}
