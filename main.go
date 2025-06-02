package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	config = flag.String("c", "sources.yml", "config file")
	output = flag.String("o", "blocklist.txt", "output hosts file")

	client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
		},
	}
)

func main() {
	flag.Parse()

	data, err := os.ReadFile(*config)
	if err != nil {
		log.Fatal(err)
	}
	var conf Config
	if err = yaml.Unmarshal(data, &conf); err != nil {
		log.Fatal(err)
	}

	myBlocked := make(map[string]bool)
	for _, host := range conf.Local.Block {
		myBlocked[host] = true
	}
	log.Printf("Found %d hosts in local block list.\n", len(myBlocked))

	myAllowed := make(map[string]bool)
	for _, host := range conf.Local.Safe {
		if _, ok := myBlocked[host]; !ok {
			myAllowed[host] = true
		}
	}
	log.Printf("Found %d hosts in local safe list.\n", len(myAllowed))

	allowed := conf.AllowLists.Get()
	log.Printf("Found %d hosts in safe lists.\n", len(allowed))
	for host := range myAllowed {
		allowed[host] = true
	}

	blocked := conf.DenyLists.Get()
	noBlocked := len(blocked)
	if len(conf.DenyLists) > 0 && noBlocked == 0 {
		log.Fatal("Failed to retrieve block lists' entries.")
	}
	log.Printf("Found %d hosts in block lists.\n", noBlocked)
	for host := range allowed {
		delete(blocked, host)
	}
	if noBlocked -= len(blocked); noBlocked > 0 {
		log.Printf("Removed %d safe hosts from blocked hosts.\n", noBlocked)
	}

	var hosts []string
	for host := range blocked {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	log.Printf("Total No. hosts allowed: %d\n", len(allowed))
	log.Printf("Total No. hosts blocked: %d\n", len(hosts))

	out, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	writer := bufio.NewWriter(out)
	for _, host := range hosts {
		writer.WriteString(fmt.Sprintf("0.0.0.0\t%s\n", host))
	}
	if err = writer.Flush(); err != nil {
		log.Fatal(err)
	}
}
