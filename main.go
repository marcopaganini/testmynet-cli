// testmynet_cli - CLI based network bandwidth tester using testmy.net.
//
// See instructions in the README.md file that accompanies this program.
//
// (C) by Marco Paganini <paganini AT paganini DOT net>

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/marcopaganini/logger"
)

const (
	stateFile          = ".testmynet-cli.state"
	minDurationMinutes = 15
)

var (
	// Generic logging object
	log *logger.Logger
)

// download retrieves test data from test servers and returns the number of
// bytes effectively read and the time it took to read those bytes.
func download(server string, datasize int, dryrun bool) (int64, time.Duration, error) {
	uri := fmt.Sprintf("%s/dl-%d", server, datasize)
	log.Verbosef(1, "Starting download from %q\n", uri)

	res, err := http.Get(uri)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	// Default values, used if we're doing a dry-run.
	written := int64(1e6)
	duration := time.Duration(8 * time.Second)

	if !dryrun {
		// Timed download
		tstart := time.Now()

		written, err = io.Copy(ioutil.Discard, res.Body)
		if err != nil {
			return 0, 0, err
		}
		duration = time.Since(tstart)
	}

	log.Verbosef(1, "%d bytes downloaded in %s\n", written, duration)
	return written, duration, nil
}

// homeDir returns the user's home directory or an error if the variable HOME
// is not set, or os.user fails, or the directory cannot be found.
func homeDir() (string, error) {
	// Get home directory from the HOME environment variable first.
	home := os.Getenv("HOME")
	if home == "" {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("unable to get information for current user: %q", err)
		}
		home = usr.HomeDir
	}
	_, err := os.Stat(home)
	if os.IsNotExist(err) || !os.ModeType.IsDir() {
		return "", fmt.Errorf("home directory %q must exist and be a directory", home)
	}
	// Other errors than file not found.
	if err != nil {
		return "", err
	}
	return home, nil
}

// overloadProtection returns an error if it's been called less than
// "minDuration" ago. If not, it creates a file under the user directory
// and writes the current timestamp there for future use.
func overloadProtection(stateFile string, minDuration time.Duration) error {
	home, err := homeDir()
	if err != nil {
		return err
	}
	fname := filepath.Join(home, stateFile)
	log.Verbosef(1, "Reading state file: %q\n", fname)

	// If the file exists, we read it and make sure the current time
	// is more than minDuration ahead of the time saved in the file.
	buf, err := ioutil.ReadFile(fname)
	switch {
	// No errors
	case err == nil:
		tim, err := time.Parse(time.UnixDate, string(buf))
		if err != nil {
			return err
		}
		since := time.Since(tim)
		log.Verbosef(1, "Last timestamp: %s, minimum interval: %s, elapsed: %s\n", tim, minDuration, since)
		if since < minDuration {
			return fmt.Errorf("program ran less than %s ago (%s)", minDuration, since)
		}
	// Any errors other than non-existing file
	case !os.IsNotExist(err):
		return err
	}

	// Rewrite current time
	tim := time.Now()
	tformat := tim.Format(time.UnixDate)
	log.Verbosef(1, "Re-writing current time (%s) to state file\n", tformat)

	if err = ioutil.WriteFile(fname, []byte(tformat), 0644); err != nil {
		return err
	}
	return nil
}

func main() {
	log = logger.New("")
	opt := &cmdLineOpts{}

	// Parse command line flags and read config file.
	if err := opt.parseFlags(); err != nil {
		log.Fatalf("Error: %s\n", err)
	}

	// Set verbose level
	verbose := int(opt.verbose)
	if verbose > 0 {
		log.SetVerboseLevel(verbose)
	}

	bytes, duration, err := download(opt.server, opt.datasize, opt.dryrun)
	if err != nil {
		log.Fatalf("Error downloading data from %s: %v\n", opt.server, err)
	}

	// Don't overload testmy.net (unless force is set).
	if !opt.force {
		if err := overloadProtection(stateFile, time.Duration(minDurationMinutes*time.Minute)); err != nil {
			log.Fatalf("Error: %s\n", err)
		}
	}

	// Calculate bandwidth and print.
	bw := (float64(bytes) * 8 / duration.Seconds()) / 1e6
	if opt.csv {
		fmt.Printf("%s,%d,%.2f,%.3f\n", opt.server, bytes, duration.Seconds(), bw)
	} else {
		fmt.Printf("Downloaded %d bytes from %s in %s. Bandwidth = %.3fMbps\n",
			bytes, opt.server, duration, bw)
	}
}
