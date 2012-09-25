package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/couchbaselabs/go-couchbase"

	"github.com/couchbaselabs/cbfs/config"
)

var workers = flag.Int("workers", 4, "Number of upload workers")
var couchbaseServer = flag.String("couchbase", "", "Couchbase URL")
var couchbaseBucket = flag.String("bucket", "default", "Couchbase bucket")
var revs = flag.Int("revs", 0, "Number of old revisions to keep (-1 == all)")

var cb *couchbase.Bucket

var commands = map[string]struct {
	nargs  int
	f      func(args []string)
	argstr string
}{
	"upload":  {-2, uploadCommand, "[opts] /src/dir http://cbfs:8484/path/"},
	"ls":      {1, lsCommand, "http://cbfs:8484/some/path"},
	"rm":      {-1, rmCommand, "[-r] [-v] http://cbfs:8484/some/path"},
	"getconf": {0, getConfCommand, ""},
	"setconf": {2, setConfCommand, "prop value"},
}

func init() {
	log.SetFlags(log.Lmicroseconds)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage of %s [-flags] cmd cmdargs\n",
			os.Args[0])

		fmt.Fprintf(os.Stderr, "\nCommands:\n")

		for k, v := range commands {
			fmt.Fprintf(os.Stderr, "  %s %s\n", k, v.argstr)
		}

		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

}

func getConfCommand(args []string) {
	if cb == nil {
		log.Fatalf("No couchbase bucket specified")
	}
	conf := cbfsconfig.DefaultConfig()
	err := conf.RetrieveConfig(cb)
	if err != nil {
		log.Printf("Error getting config: %v", err)
		log.Printf("Using default, as shown below:")
	}

	conf.Dump(os.Stdout)
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("Unable to parse duration: %v", err)
	}
	return d
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Error parsing int: %v", err)
	}
	return i
}

func setConfCommand(args []string) {
	if cb == nil {
		log.Fatalf("No couchbase bucket specified")
	}
	conf := cbfsconfig.DefaultConfig()
	err := conf.RetrieveConfig(cb)
	if err != nil {
		log.Printf("Error getting config: %v, using default", err)
	}

	switch args[0] {
	default:
		log.Fatalf("Unhandled property: %v (try running getconf)",
			args[0])
	case "gcfreq":
		conf.GCFreq = parseDuration(args[1])
	case "gclimit":
		conf.GCLimit = parseInt(args[1])
	case "hash":
		conf.Hash = args[1]
	case "hbfreq":
		conf.HeartbeatFreq = parseDuration(args[1])
	case "minrepl":
		conf.MinReplicas = parseInt(args[1])
	case "maxrepl":
		conf.MaxReplicas = parseInt(args[1])
	case "cleanCount":
		conf.NodeCleanCount = parseInt(args[1])
	case "reconcileFreq":
		conf.ReconcileFreq = parseDuration(args[1])
	case "nodeCheckFreq":
		conf.StaleNodeCheckFreq = parseDuration(args[1])
	case "staleLimit":
		conf.StaleNodeLimit = parseDuration(args[1])
	case "underReplicaCheckFreq":
		conf.UnderReplicaCheckFreq = parseDuration(args[1])
	case "overReplicaCheckFreq":
		conf.OverReplicaCheckFreq = parseDuration(args[1])
	}

	err = conf.StoreConfig(cb)
	if err != nil {
		log.Fatalf("Error updating config: %v", err)
	}
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
	}

	if *couchbaseServer != "" {
		var err error
		cb, err = couchbase.GetBucket(*couchbaseServer,
			"default", *couchbaseBucket)
		if err != nil {
			log.Fatalf("Error connecting to couchbase: %v", err)
		}
	}

	cmdName := flag.Arg(0)
	cmd, ok := commands[cmdName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %v\n", cmdName)
		flag.Usage()
	}
	if cmd.nargs < 0 {
		reqargs := -cmd.nargs
		if flag.NArg()-1 < reqargs {
			fmt.Fprintf(os.Stderr, "Incorrect arguments for %v\n", cmdName)
			flag.Usage()
		}
	} else {
		if flag.NArg()-1 != cmd.nargs {
			fmt.Fprintf(os.Stderr, "Incorrect arguments for %v\n", cmdName)
			flag.Usage()
		}
	}

	cmd.f(flag.Args()[1:])
}
