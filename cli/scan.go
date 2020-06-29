package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// Scan service structure
type Scan struct {
	Service

	hostname string
}

// NewScanCmd returns a new Scan command
func NewScanCmd() *Scan {
	// Define the Cobra command
	methodCmd := &cobra.Command{
		Use:     "scan",
		Short:   "Scan a hostname to get the main CDN provider",
		Long:    "Scan a hostname to get the main CDN provider",
		Example: "cdn-scanner scan --hostname <host name>",
	}

	// Create the Scan instance
	tmp := &Scan{}
	// Add the Cobra command to it
	tmp.cmd = methodCmd

	// Define the command specific flags
	methodCmd.Flags().StringVarP(&tmp.hostname, "hostname", "", "", "Hostname to scan for CDN")

	return tmp
}

// GetCmd returns the defined Cobra command
func (sc *Scan) GetCmd() *cobra.Command {
	return sc.cmd
}

// Setup the command execution
func (sc *Scan) Setup() error {
	// Validate input flags
	if sc.hostname == "" {
		return fmt.Errorf("error: hostname flag is empty")
	}

	return nil
}

// Execute the command
func (sc *Scan) Execute() error {
	// Lookup for NS records for the given hostname
	nsLookup, err := net.LookupNS(sc.hostname)
	if err != nil {
		return fmt.Errorf("error: could not lookup NS records: %w", err)
	}

	// Read & unmarshal the known CDN NS records JSON into map
	var nsRecords map[string][]string
	if err := readJSONtoObject("resources/ns_records.json", &nsRecords); err != nil {
		return err
	}

	// Convert the NS lookup results into a string slice of hosts
	nsHosts := make([]string, len(nsLookup))
	for idx, nsRecord := range nsLookup {
		nsHosts[idx] = nsRecord.Host
	}

	// Join the string slice of hosts into a single string
	nsHostsString := strings.Join(nsHosts, "::")

	// Go through the map of CDN providers & hosts to find a match from the NS hosts
	for cdnProvider, knownHosts := range nsRecords {
		// Use regex to simplify the matching
		regx := regexp.MustCompile(strings.Join(knownHosts, "|"))

		if regx.MatchString(nsHostsString) {
			fmt.Println("CDN FOUND: ", cdnProvider)
			return nil
		}
	}

	// Read & unmarshal the known CDN header keys JSON into map
	var headerKeys map[string][]string
	if err := readJSONtoObject("resources/header_keys.json", &headerKeys); err != nil {
		return err
	}

	// Make an HTTPS request to the host in order to get a response
	httpsResp, err := http.Get("https://" + sc.hostname)
	if err != nil {
		return fmt.Errorf("error: could not HTTPS GET from host: %w", err)
	}

	// Check the headers from the HTTPS response for a CDN provider
	if cdnProvider := getProviderFromHeaders(httpsResp, headerKeys); cdnProvider != "" {
		fmt.Println("CDN FOUND: ", cdnProvider)
		return nil
	}

	// Make an HTTP request to the host in order to get a response
	httpResp, err := http.Get("http://" + sc.hostname)
	if err != nil {
		return fmt.Errorf("error: could not HTTP GET from host: %w", err)
	}

	// Check the headers from the HTTP response for a CDN provider
	if cdnProvider := getProviderFromHeaders(httpResp, headerKeys); cdnProvider != "" {
		fmt.Println("CDN FOUND: ", cdnProvider)
		return nil
	}

	// No CDN provider found
	fmt.Println("CDN NOT FOUND!")
	return nil
}

// getProviderFromHeaders will check the headers from an HTTP(S) response matching the known keys
func getProviderFromHeaders(resp *http.Response, headerKeys map[string][]string) string {
	// Check the `server` header for a value
	keyServer := resp.Header.Get("server")
	if keyServer != "" {
		return strings.ToUpper(keyServer)
	}

	// Check the other known keys
	for cdnProvider, knownKeys := range headerKeys {
		for _, knownKey := range knownKeys {
			keyQuery := resp.Header.Get(knownKey)
			if keyQuery != "" {
				return cdnProvider
			}
		}
	}

	// Nothing found
	return ""
}

// readJSONtoObject reads a given json file name and unmarshals it into a given map
func readJSONtoObject(fileName string, dest *map[string][]string) error {
	if file, err := ioutil.ReadFile(fileName); err != nil {
		return fmt.Errorf("error: could not read `%s` file: %w", fileName, err)
	} else if err := json.Unmarshal(file, dest); err != nil {
		return fmt.Errorf("error: could unmarshal `%s`: %w", fileName, err)
	}

	return nil
}
