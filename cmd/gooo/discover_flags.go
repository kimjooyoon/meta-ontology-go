package main

import "strings"

func parseDiscoveryOptions(args []string) (flags map[string]string, filename string, help, valid bool) {
	flags = map[string]string{"--query": "", "--domain-contract": "", "--generation": "", "--execute-cases": "", "--go-bin": ""}
	for index := 0; index < len(args); index++ {
		flag := args[index]
		if flag == "--help" || flag == "-h" {
			return flags, filename, true, true
		}
		if _, exists := flags[flag]; exists {
			if flags[flag] != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" || strings.HasPrefix(args[index+1], "--") {
				return flags, filename, false, false
			}
			index++
			flags[flag] = args[index]
			if flag != "--execute-cases" && flag != "--go-bin" {
				flags[flag] = strings.TrimSpace(flags[flag])
			}
			continue
		}
		if strings.HasPrefix(flag, "-") || filename != "" {
			return flags, filename, false, false
		}
		filename = flag
	}
	valid = flags["--query"] != "" && filename != "" &&
		(flags["--execute-cases"] == "" || flags["--generation"] != "") &&
		(flags["--go-bin"] == "" || flags["--execute-cases"] != "")
	return flags, filename, false, valid
}
