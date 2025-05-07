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
				var domain string
				line := normalizeEntry(scanner.Text())
				if line == "" {
					continue
				}
				switch list.Type {
				case Hosts:
					if !strings.HasPrefix(line, "0.0.0.0") && !strings.HasPrefix(line, "127.0.0.1") {
						continue
					}
					var found bool
					if _, domain, found = strings.Cut(line, " "); found {
						switch domain {
						case "localhost", "localhost.localdomain", "local", "":
							continue
						}
					}
				case Domains:
					if strings.Contains(line, " ") {
						continue
					}
					domain = line
				}
				if _, ok := dns.IsDomainName(domain); !ok {
					continue
				}
				res[domain] = true
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

func normalizeEntry(line string) string {
	line, _, _ = strings.Cut(line, "#")
	return strings.TrimSpace(strings.ToLower(strings.ReplaceAll(line, "\t", " ")))
}
