package domainscan

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/miekg/dns"
	"github.com/remeh/sizedwaitgroup"
	"github.com/rs/xid"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type Massdns struct {
	Config  Config
	Result  Result
	tempDir string
}

/*
 forked from https://github.com/projectdiscovery/shuffledns
*/

// Client is a client for running massdns on a target
type Client struct {
	config ClientConfig

	wildcardIPMap   map[string]struct{}
	wildcardIPMutex *sync.RWMutex

	wildcardResolver *Resolver
}

// ClientConfig contains configuration options for the massdns client
type ClientConfig struct {
	// Domain is the domain specified for enumeration
	Domain string
	// Retries is the nmber of retries for dns
	Retries int
	// MassdnsPath is the path to the binary
	MassdnsPath string
	// Threads is the hashmap size for massdns
	Threads int
	// InputFile is the file to use for massdns input
	InputFile string
	// ResolversFile is the file with the resolvers
	ResolversFile string
	// TempDir is a temporary directory for storing massdns misc files
	TempDir string
	// OutputFile is the file to use for massdns output
	OutputFile string
	// WildcardsThreads is the number of wildcards concurrent threads
	WildcardsThreads int
	// MassdnsRaw perform wildcards filtering from an existing massdns output file
	MassdnsRaw string
	// StrictWildcard controls whether the wildcard check should be performed on each result
	StrictWildcard bool
}

// Store is a storage for ip based wildcard removal
type Store struct {
	IP map[string]*IPMeta
}

// IPMeta contains meta-information about a single
// IP address found during enumeration.
type IPMeta struct {
	// we store also the ip itself as we will need it later for filtering
	IP string
	// Hostnames contains the list of hostnames for the IP
	Hostnames map[string]struct{}
	// Counter is the number of times the same ip was found for hosts
	Counter int
}

// Resolver represents a dns resolver for removing wildcards
type Resolver struct {
	// servers contains the dns servers to use
	servers []string
	// serversIndex contains the current pointer to servers.
	// All the DNS resolvers are used in a round robin fashion.
	serversIndex int32
	// domain is the domain to perform enumeration on
	domain string
	// maxRetries is the maximum number of retries allowed
	maxRetries int
}

// excellentResolvers contains some resolvers used in dns verification step
var excellentResolvers = []string{
	"1.1.1.1",
	"1.0.0.1",
	"8.8.8.8",
	"8.8.4.4",
}

// NewMassdns 创建Massdns对象
func NewMassdns(config Config) *Massdns {
	return &Massdns{Config: config}
}

// Do 执行Massdns任务
func (m *Massdns) Do() {
	dir, err := ioutil.TempDir(filepath.Join(conf.GetRootPath(), "thirdparty/massdns/temp"), utils.GetRandomString2(8))
	if err != nil {
		return
	}
	m.tempDir = dir
	m.Result.DomainResult = make(map[string]*DomainResult)
	swg := sizedwaitgroup.New(massdnsThreadNumber)

	for _, line := range strings.Split(m.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" || utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
			continue
		}
		swg.Add()
		go func(d string) {
			m.processDomain(d)
			swg.Done()
		}(domain)
	}
	swg.Wait()
	m.Close()
}

// parseResult 解析子域名枚举结果文件
func (m *Massdns) parseResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(content), "\n") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		if !m.Result.HasDomain(domain) {
			m.Result.SetDomain(domain)
		}
	}
}

// Close 执行完成后的临时目录与文件的清除
func (m *Massdns) Close() {
	os.RemoveAll(m.tempDir)
}

// processDomain processes the bruteforce for a domain using a wordlist
func (m *Massdns) processDomain(domain string) {
	tempOutputFile := utils.GetTempPathFileName()
	defer os.Remove(tempOutputFile)

	resolveFile := filepath.Join(m.tempDir, xid.New().String())
	file, err := os.Create(resolveFile)
	if err != nil {
		logging.RuntimeLog.Errorf("Could not create bruteforce list (%s): %s\n", m.tempDir, err)
		return
	}
	writer := bufio.NewWriter(file)

	// Read the input wordlist for bruteforce generation
	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/dict", conf.Nemo.Domainscan.Wordlist))
	if err != nil {
		logging.RuntimeLog.Errorf("Could not read bruteforce wordlist: %s\n", err)
		file.Close()
		return
	}

	//gologger.Infof("Started generating bruteforce permutation\n")

	//now := time.Now()
	// Create permutation for domain with wordlist
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		// RFC4343 - case insensitive domain
		text := strings.ToLower(scanner.Text())
		if text == "" {
			continue
		}
		writer.WriteString(text + "." + domain + "\n")
	}
	writer.Flush()
	inputFile.Close()
	file.Close()

	//gologger.Infof("Generating permutations took %s\n", time.Now().Sub(now))

	// Run the actual massdns enumeration process
	m.runMassdns(domain, resolveFile, tempOutputFile)
	m.parseResult(tempOutputFile)
}

// runMassdns runs the massdns tool on the list of inputs
func (m *Massdns) runMassdns(domain, inputFile, outputFile string) {
	cc := ClientConfig{
		Domain:           domain,
		MassdnsPath:      filepath.Join(conf.GetRootPath(), "thirdparty/massdns", "massdns_darwin_amd64"),
		Threads:          conf.Nemo.Domainscan.MassdnsThreads,
		InputFile:        inputFile,
		ResolversFile:    filepath.Join(conf.GetRootPath(), "thirdparty/dict",conf.Nemo.Domainscan.Resolver),
		TempDir:          m.tempDir,
		OutputFile:       outputFile,
		Retries:          5,
		WildcardsThreads: 25,
	}
	if runtime.GOOS == "linux" {
		cc.MassdnsPath = filepath.Join(conf.GetRootPath(), "thirdparty/massdns", "massdns_linux_amd64")
	}
	massdns, err := newMassdnsClient(cc)
	if err != nil {
		logging.RuntimeLog.Errorf("Could not create massdns client: %s\n", err)
		return
	}

	err = massdns.Process()
	if err != nil {
		logging.RuntimeLog.Errorf("Could not run massdns: %s\n", err)
	}
	//logging.RuntimeLog.Infof("Finished resolving. Hack the Planet!\n")
}

// newMassdnsClient returns a new massdns client for running enumeration
// on a target.
func newMassdnsClient(config ClientConfig) (*Client, error) {
	// Create a resolver and load resolverrs from list
	resolver, err := NewResolver(config.Domain, config.Retries)
	if err != nil {
		return nil, err
	}

	resolver.AddServersFromList(excellentResolvers)

	return &Client{
		config: config,

		wildcardIPMap:    make(map[string]struct{}),
		wildcardIPMutex:  &sync.RWMutex{},
		wildcardResolver: resolver,
	}, nil
}

// Process runs the actual enumeration process returning a file
func (c *Client) Process() error {
	// Create a store for storing ip metadata
	shstore := NewStore()
	defer shstore.Close()

	// Set the correct target file
	massDNSOutput := path.Join(c.config.TempDir, xid.New().String())
	if c.config.MassdnsRaw != "" {
		massDNSOutput = c.config.MassdnsRaw
	}

	// Check if we need to run massdns
	if c.config.MassdnsRaw == "" {
		// Create a temporary file for the massdns output
		//gologger.Infof("Creating temporary massdns output file: %s\n", massDNSOutput)
		err := c.runMassDNS(massDNSOutput, shstore)
		if err != nil {
			return fmt.Errorf("could not execute massdns: %w", err)
		}
	}

	//gologger.Infof("Started parsing massdns output\n")

	err := c.parseMassDNSOutput(massDNSOutput, shstore)
	if err != nil {
		return fmt.Errorf("could not parse massdns output: %w", err)
	}

	//gologger.Infof("Massdns output parsing compeleted\n")

	// Perform wildcard filtering only if domain name has been specified
	if c.config.Domain != "" {
		//gologger.Infof("Started removing wildcards records\n")
		err = c.filterWildcards(shstore)
		if err != nil {
			return fmt.Errorf("could not parse massdns output: %w", err)
		}
		//gologger.Infof("Wildcard removal completed\n")
	}

	//gologger.Infof("Finished enumeration, started writing output\n")

	// Write the final elaborated list out
	return c.writeOutput(shstore)
}

func (c *Client) runMassDNS(output string, store *Store) error {
	//if c.config.Domain != "" {
	//	gologger.Infof("Executing massdns on %s\n", c.config.Domain)
	//} else {
	//	gologger.Infof("Executing massdns\n")
	//}
	//now := time.Now()
	// Run the command on a temp file and wait for the output
	cmd := exec.Command(c.config.MassdnsPath, []string{"-r", c.config.ResolversFile, "-o", "Snl", "-t", "A", c.config.InputFile, "-w", output, "-s", strconv.Itoa(c.config.Threads)}...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not execute massdns: %w\ndetailed error: %s", err, stderr.String())
	}
	//gologger.Infof("Massdns execution took %s\n", time.Now().Sub(now))
	return nil
}

func (c *Client) parseMassDNSOutput(output string, store *Store) error {
	massdnsOutput, err := os.Open(output)
	if err != nil {
		return fmt.Errorf("could not open massdns output file: %w", err)
	}
	defer massdnsOutput.Close()

	// at first we need the full structure in memory to elaborate it in parallell
	err = Parse(massdnsOutput, func(domain string, ip []string) {
		for _, ip := range ip {
			// Check if ip exists in the store. If not,
			// add the ip to the map and continue with the next ip.
			if !store.Exists(ip) {
				store.New(ip, domain)
				continue
			}

			// Get the Domain meta-information from the store.
			record := store.Get(ip)

			// Put the new hostname and increment the counter by 1.
			record.Hostnames[domain] = struct{}{}
			record.Counter++
		}
	})

	if err != nil {
		return fmt.Errorf("could not parse massdns output: %w", err)
	}

	return nil
}

func (c *Client) filterWildcards(st *Store) error {
	// Start to work in parallel on wildcards
	wildcardWg := sizedwaitgroup.New(c.config.WildcardsThreads)

	for _, record := range st.IP {
		// We've stumbled upon a wildcard, just ignore it.
		c.wildcardIPMutex.Lock()
		if _, ok := c.wildcardIPMap[record.IP]; ok {
			c.wildcardIPMutex.Unlock()
			continue
		}
		c.wildcardIPMutex.Unlock()

		// Perform wildcard detection on the ip, if an Domain is found in the wildcard
		// we add it to the wildcard map so that further runs don't require such filtering again.
		if record.Counter >= 5 || c.config.StrictWildcard {
			wildcardWg.Add()
			go func(record *IPMeta) {
				defer wildcardWg.Done()

				for host := range record.Hostnames {
					isWildcard, ips := c.wildcardResolver.LookupHost(host)
					if len(ips) > 0 {
						c.wildcardIPMutex.Lock()
						for ip := range ips {
							// we add the single ip to the wildcard list
							c.wildcardIPMap[ip] = struct{}{}
						}
						c.wildcardIPMutex.Unlock()
					}

					if isWildcard {
						c.wildcardIPMutex.Lock()
						// we also mark the original ip as wildcard, since at least once it resolved to this host
						c.wildcardIPMap[record.IP] = struct{}{}
						c.wildcardIPMutex.Unlock()
						break
					}
				}
			}(record)
		}
	}

	wildcardWg.Wait()

	// drop all wildcard from the store
	for wildcardIP := range c.wildcardIPMap {
		st.Delete(wildcardIP)
	}

	return nil
}

func (c *Client) writeOutput(store *Store) error {
	// Write the unique deduplicated output to the file or stdout
	// depending on what the user has asked.
	var output *os.File
	var w *bufio.Writer
	var err error

	if c.config.OutputFile != "" {
		output, err = os.Create(c.config.OutputFile)
		if err != nil {
			return fmt.Errorf("could not create massdns output file: %v", err)
		}
		w = bufio.NewWriter(output)
	}
	buffer := &strings.Builder{}

	uniqueMap := make(map[string]struct{})

	for _, record := range store.IP {
		for hostname := range record.Hostnames {
			// Skip if we already printed this subdomain once
			if _, ok := uniqueMap[hostname]; ok {
				continue
			}
			uniqueMap[hostname] = struct{}{}

			buffer.WriteString(hostname)
			buffer.WriteString("\n")
			data := buffer.String()

			if output != nil {
				w.WriteString(data)
			}
			//gologger.Silentf("%s", data)
			logging.CLILog.Infof("%s", data)
			buffer.Reset()
		}
	}

	// Close the files and return
	if output != nil {
		w.Flush()
		output.Close()
	}
	return nil
}

// NewStore Client creates a new storage for ip based wildcard removal
func NewStore() *Store {
	return &Store{
		IP: make(map[string]*IPMeta),
	}
}

// New creates a new ip-hostname pair in the map
func (s *Store) New(ip, hostname string) {
	hostnames := make(map[string]struct{})
	hostnames[hostname] = struct{}{}
	s.IP[ip] = &IPMeta{IP: ip, Hostnames: hostnames, Counter: 1}
}

// Exists indicates if an IP exists in the map
func (s *Store) Exists(ip string) bool {
	_, ok := s.IP[ip]
	return ok
}

// Get gets the meta-information for an IP address from the map.
func (s *Store) Get(ip string) *IPMeta {
	return s.IP[ip]
}

// Delete deletes the records for an IP from store.
func (s *Store) Delete(ip string) {
	delete(s.IP, ip)
}

// Close removes all the references to arrays and releases memory to the gc
func (s *Store) Close() {
	for ip := range s.IP {
		s.IP[ip].Hostnames = nil
	}
}

// Callback is a callback function that is called by
// the parser returning the results found.
// NOTE: Callbacks are not thread safe and are blocking in nature
// and should be used as such.
type Callback func(domain string, ip []string)

// Parse parses the massdns output returning the found
// domain and ip pair to a callback function.
//
// It's a pretty hacky solution. In future, it can and should
// be rewritten to handle more edge cases and stuff.
func Parse(reader io.Reader, callback Callback) error {
	var (
		// Some boolean various needed for state management
		cnameStart bool
		nsStart    bool

		// Result variables to store the results
		domain string
		ip     []string
	)

	// Parse the input line by line and act on what the line means
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()

		// Empty line represents a seperator between DNS reply
		// due to `-o Snl` option set in massdns. Thus it can be
		// interpreted as a DNS answer header.
		//
		// If we have start of a DNS answer header, set the
		// bool state to default, and return the results to the
		// consumer via the callback.
		if text == "" {
			if domain != "" {
				cnameStart, nsStart = false, false
				callback(domain, ip)
				domain, ip = "", nil
			}
			continue
		} else {
			// Non empty line represents DNS answer section, we split on space,
			// iterate over all the parts, and write the answer to the struct.
			parts := strings.Split(text, " ")

			if len(parts) != 3 {
				continue
			}

			// Switch on the record type, deciding what to do with
			// a record based on the type of record.
			switch parts[1] {
			case "NS":
				// If we have a NS record, then set nsStart
				// which will ignore all the next records
				nsStart = true
			case "CNAME":
				// If we have a CNAME record, then the next record should be
				// the values for the CNAME record, so set the cnameStart value.
				//
				// Use the domain in the first cname field since the next fields for
				// A record may contain domain for secondary CNAME which messes
				// up recursive CNAME records.
				if !cnameStart {
					nsStart = false
					domain = strings.TrimSuffix(parts[0], ".")
					cnameStart = true
				}
			case "A":
				// If we have an A record, check if it's not after
				// an NS record. If not, append it to the ips.
				//
				// Also if we aren't inside a CNAME block, set the domain too.
				if !nsStart {
					if !cnameStart && domain == "" {
						domain = strings.TrimSuffix(parts[0], ".")
					}
					ip = append(ip, parts[2])
				}
			}
		}
		continue
	}

	// Return error if there was any.
	if err := scanner.Err(); err != nil {
		return err
	}

	// Final callback to deliver the last piece of result
	// if there's any.
	if domain != "" {
		callback(domain, ip)
	}
	return nil
}

// NewResolver initializes and creates a new resolver to find wildcards
func NewResolver(domain string, retries int) (*Resolver, error) {
	resolver := &Resolver{
		servers:      []string{},
		serversIndex: 0,
		domain:       domain,
		maxRetries:   retries,
	}
	return resolver, nil
}

// AddServersFromList adds the resolvers from a list of servers
func (w *Resolver) AddServersFromList(list []string) {
	for _, server := range list {
		w.servers = append(w.servers, server+":53")
	}
}

// AddServersFromFile adds the resolvers from a file to the list of servers
func (w *Resolver) AddServersFromFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		w.servers = append(w.servers, text+":53")
	}
	return nil
}

// LookupHost returns wildcard Domain addresses of a wildcard if it's a wildcard.
// To determine, first we split the target host by dots, create permutation
// of it's levels, check for wildcard on each one of them and if found any,
// we remove all the hosts that have this Domain from the map.
func (w *Resolver) LookupHost(host string) (bool, map[string]struct{}) {
	orig := make(map[string]struct{})
	wildcards := make(map[string]struct{})

	subdomainPart := strings.TrimSuffix(host, "."+w.domain)
	subdomainTokens := strings.Split(subdomainPart, ".")

	// Build an array by preallocating a slice of a length
	// and create the wildcard generation prefix.
	// We use a rand prefix at the beginning like %rand%.domain.tld
	// A permutation is generated for each level of the subdomain.
	var hosts []string
	hosts = append(hosts, host)
	hosts = append(hosts, xid.New().String()+"."+w.domain)

	for i := 0; i < len(subdomainTokens); i++ {
		newhost := xid.New().String() + "." + strings.Join(subdomainTokens[i:], ".") + "." + w.domain
		hosts = append(hosts, newhost)
	}

	// Iterate over all the hosts generated for rand.
	for _, h := range hosts {
		// Round-robin over all the dns servers we have.
		serverIndex := atomic.LoadInt32(&w.serversIndex)
		if w.serversIndex >= int32(len(w.servers)-1) {
			atomic.StoreInt32(&w.serversIndex, 0)
			serverIndex = 0
		}
		resolver := w.servers[serverIndex]
		atomic.AddInt32(&w.serversIndex, 1)

		var retryCount int

	retry:

		// Create a dns message and send it to the server
		m := new(dns.Msg)
		m.Id = dns.Id()
		m.RecursionDesired = true
		m.Question = make([]dns.Question, 1)
		question := dns.Fqdn(h)
		m.Question[0] = dns.Question{
			Name:   question,
			Qtype:  dns.TypeA,
			Qclass: dns.ClassINET,
		}
		in, err := dns.Exchange(m, resolver)
		if err != nil {
			if retryCount < w.maxRetries {
				retryCount++
				goto retry
			}
			// Skip the current host if there are no more retries
			retryCount = 0
			continue
		}

		// Skip the current host since we can't resolve it
		if in != nil && in.Rcode != dns.RcodeSuccess {
			continue
		}

		// Get all the records and add them to the wildcard map
		for _, record := range in.Answer {
			if t, ok := record.(*dns.A); ok {
				r := t.A.String()

				if host == h {
					orig[r] = struct{}{}
					continue
				}

				if _, ok := wildcards[r]; !ok {
					wildcards[r] = struct{}{}
				}
			}
		}
	}

	// check if original ip are among wildcards
	for a := range orig {
		if _, ok := wildcards[a]; ok {
			return true, wildcards
		}
	}

	return false, wildcards
}
