package main

import (
	"bufio"
	"log"
	"strings"

	"github.com/miekg/dns"
)

type Format string

const (
	Domains Format = "domains"
	Hosts   Format = "hosts"
)

type List struct {
	Name    string   `yaml:"name"`
	URL     string   `yaml:"url"`
	Repo    string   `yaml:"repository"`
	Sources []string `yaml:"sources"`
	Type    Format   `yaml:"type"`
}

type Lists []List

type LocalLists struct {
	Safe  []string `yaml:"safe"`
	Block []string `yaml:"block"`
}

type Config struct {
	AllowLists Lists      `yaml:"allow"`
	DenyLists  Lists      `yaml:"deny"`
	Local      LocalLists `yaml:"local"`
}

func (lists Lists) Get() map[string]bool {
	res := make(map[string]bool)
	for _, list := range lists {
		log.Printf("Processing %q...\n", list.Name)
		for _, url := range list.Sources {
			resp, err := client.Get(url)
			if err != nil {
				log.Printf("ERROR: %v\n", err)
				continue
			}
			defer resp.Body.Close()
			var entries int
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.TrimSpace(strings.ToLower(scanner.Text()))
				switch {
				case line == "",
					strings.HasPrefix(line, "#"),
					list.Type == Domains && strings.Contains(line, " "):
					continue
				default:
					if list.Type == Hosts {
						if strings.HasPrefix(line, "0.0.0.0") || strings.HasPrefix(line, "127.0.0.1") {
							switch line = strings.Split(line, " ")[1]; line {
							case "localhost", "localhost.localdomain":
								continue
							}
						} else {
							continue
						}
					}
					if _, ok := dns.IsDomainName(line); !ok {
						continue
					}
				}
				res[line] = true
				entries++
			}
			if err := scanner.Err(); err != nil {
				log.Printf("ERROR: %v\n", err)
			}
			log.Printf("Found %d entries in %s\n", entries, url)
		}
	}
	return res
}
