// This file is part of testmynet_cli (http://github.com/marcopaganini/testmynet_cli)
// See instructions in the README.md file that accompanies this program.
// (C) by Marco Paganini <paganini AT paganini DOT net>

package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

const (
	// Flag defaults.
	defaultLocation = "ca"
	defaultDataSize = 10240
	tmnDomain       = "testmy.net"
)

type multiLevelInt int

type cmdLineOpts struct {
	csv      bool
	datasize int
	dryrun   bool
	force    bool
	location string
	server   string
	verbose  multiLevelInt
}

var (
	// Command line Flags.
	opt cmdLineOpts

	// TMN server locations.
	serverLocation = map[string]string{
		"au2": "Australia >> Sydney, AU",
		"ca":  "Bay Area US >> California, CA, USA",
		"co":  "Central US >> Colorado Springs, CO, USA",
		"de":  "Europe >> Frankfurt, DE",
		"fl":  "East Coast US >> Miami, FL",
		"in":  "Asia >> Bangalore, IN",
		"jp":  "Asia >> Tokyo, JP",
		"lax": "West Coast US >> Los Angeles, CA, USA",
		"ny":  "East Coast US >> New York, NY, USA",
		"sf":  "West Coast US >> San Francisco, CA, USA",
		"sg":  "Asia >> Singapore, SG",
		"tx":  "Central US >> Dallas, TX, USA",
		"uk":  "Europe >> London, GB",
	}
)

// Definitions for the custom flag type multiLevelInt.

// Return the string representation of the flag.
// The String method's output will be used in diagnostics.
func (m *multiLevelInt) String() string {
	return fmt.Sprint(*m)
}

// Increase the value of multiLevelInt. This accepts multiple values
// and sets the variable to the number of times those values appear in
// the command-line. Useful for "verbose" and "Debug" levels.
func (m *multiLevelInt) Set(_ string) error {
	*m++
	return nil
}

// Behave as a bool (i.e. no arguments).
func (m *multiLevelInt) IsBoolFlag() bool {
	return true
}

// parseFlags parses the command line and set the global opt variable. Return
// error if the basic sanity checking of flags fails.
func (x *cmdLineOpts) parseFlags() error {
	flag.BoolVar(&x.csv, "csv", false, "Output results in csv")
	flag.StringVar(&x.server, "server", "", "TestMyNet server (Overrides location)")
	flag.StringVar(&x.location, "location", defaultLocation, "TestMyNet location")
	flag.IntVar(&x.datasize, "size", defaultDataSize, "Test size in KBytes")
	flag.BoolVar(&x.dryrun, "dry-run", false, "Dry-run mode")
	flag.BoolVar(&x.force, "I-WANT-TO-GET-BANNED", false, "Allow program to hit testmy.net more often than it should.")
	flag.Var(&x.verbose, "verbose", "Verbose mode (use multiple times to increase level)")
	flag.Parse()

	// Print list of locations if location == help and exit.
	if x.location == "help" {
		fmt.Print(locationList(serverLocation))
		os.Exit(2)
	}

	// Invalid location?
	if _, ok := serverLocation[x.location]; !ok {
		return fmt.Errorf("unable to find location %q. Use \"--location help\" to see all locations", x.location)
	}

	// Fill in server with server name based on location
	// (if server was not directly specified)
	if x.server == "" {
		x.server = fmt.Sprintf("http://%s.%s", x.location, tmnDomain)
	}

	return nil
}

// locationList returns a formatted list with the location codes and
// location description.
func locationList(sloc map[string]string) string {
	keys := []string{}
	for k := range serverLocation {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	s := "Available Locations:\n"
	for ix := 0; ix < len(keys); ix++ {
		k := keys[ix]
		s = s + fmt.Sprintf("%-4.4s %s\n", k, sloc[k])
	}
	return s
}
