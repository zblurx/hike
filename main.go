package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type headers []string

type options struct {
	Url                 string
	Cookies             string
	Proxy               string
	NotCheckCert        bool
	Headers             headers
	StatusCodeBlacklist string
	OutputFile          string
	Permute             bool
	NoRequest           bool
	Csv                 bool
	NoColor             bool
}

type request struct {
	RequestURL string
	Request    *http.Request
	Client     *http.Client
}

type response struct {
	Response   *http.Response
	RequestURL string
}

var o options

var colorReset string = "\033[0m"
var colorRed string = "\033[31m"
var colorGreen string = "\033[32m"
var colorMajenta string = "\033[35m"
var colorCyan string = "\033[36m"

func colorString(color string, output string) string {
	return color + output + colorReset
}

func isError(e error) bool {
	if e != nil {
		fmt.Println(e.Error())
		return true
	}
	return false
}

func prepareClient() *http.Client {

	var proxyClient func(*http.Request) (*url.URL, error)
	if o.Proxy == "" {
		proxyClient = http.ProxyFromEnvironment
	} else {
		tmp, _ := url.Parse(o.Proxy)
		proxyClient = http.ProxyURL(tmp)
	}

	transport := &http.Transport{
		Proxy:               proxyClient,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: o.NotCheckCert},
	}

	client := &http.Client{
		Transport: transport,
	}

	return client
}

func prepareRequest(requestURL string, host string) *http.Request {

	req, err := http.NewRequest("GET", requestURL, nil)
	if isError(err) {
		os.Exit(1)
	}

	if o.Cookies != "" {
		req.Header.Set("Cookie", o.Cookies)
	}

	if host != "" {
		req.Host = host
	}

	for _, header := range o.Headers {
		h := strings.Split(header, ":")
		req.Header.Set(h[0], h[1])
	}

	return req
}

func (i *headers) String() string {
	var rep string
	for _, e := range *i {
		rep += e
	}
	return rep
}

func (i *headers) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func setHostHeaderIfExists() (host string) {
	for _, header := range o.Headers {
		h := strings.Split(header, ":")
		if len(h) != 2 {
			fmt.Printf("Error in headers declaration: %s\n", header)
		}
		if h[0] == "Host" {
			host = h[1]
		}
	}
	return
}

func getResponseFromURL(r request) *http.Response {
	response, err := r.Client.Do(r.Request)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return response
}

func removeElementFromSlice(element string, slice []string) []string {
	if element == slice[0] {
		return slice[1:]
	}
	v := []string{slice[0]}
	return append(v, removeElementFromSlice(element, slice[1:])...)
}

func permute(queue []string) (permutations [][]string) {
	for _, e := range queue {
		for _, f := range permute(removeElementFromSlice(e, queue)) {
			permutations = append(permutations, append([]string{e}, f...))
		}
		permutations = append(permutations, []string{e})
	}
	return
}

func getPermutation(baseUrl string, path []string) (urls []string) {
	permutations := permute(path)
	for _, e := range permutations {
		tmpPath := strings.Join(e, "/")
		tmpUrl := baseUrl + "/" + tmpPath
		urls = append(urls, tmpUrl)
	}
	urls = append(urls, baseUrl)
	return urls
}

func splitUrl(url_input string) (string, []string) {
	u, err := url.Parse(url_input)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	path := strings.Split(u.Path, "/")
	path = path[1:]

	clearUrl := u.Scheme + "://"
	if u.User.String() != "" {
		clearUrl += u.User.String() + "@"
	}
	clearUrl += u.Host
	if u.Port() != "" {
		clearUrl += ":" + u.Port()
	}
	return clearUrl, path
}

func hikePath(baseUrl string, path []string) (urls []string) {
	for i := len(path); i > 0; i-- {
		tmpPath := strings.Join(path[:i], "/")
		tmpUrl := baseUrl + "/" + tmpPath
		urls = append(urls, tmpUrl)
	}
	urls = append(urls, baseUrl)

	return urls
}

func getTitle(resp *http.Response) string {
	htmlTokens := html.NewTokenizer(resp.Body)
	var isTitle bool
	for {
		tt := htmlTokens.Next()
		switch tt {
		case html.ErrorToken:
			return ""
		case html.StartTagToken:
			t := htmlTokens.Token()
			isTitle = t.Data == "title"
		case html.TextToken:
			t := htmlTokens.Token()
			if isTitle {
				return t.Data
			}
			isTitle = false
		}
	}
}

func printResponse(response *http.Response, url string, file *os.File, formatString string, noColor bool, Csv bool) {
	var statusCode string

	if noColor || Csv {
		colorRed = "\033[0m"
		colorGreen = "\033[0m"
		colorMajenta = "\033[0m"
		colorCyan = "\033[0m"
	}
	if strings.Contains(response.Status, "200") {
		statusCode = colorString(colorGreen, fmt.Sprint(response.StatusCode))
	} else {
		statusCode = colorString(colorRed, fmt.Sprint(response.StatusCode))
	}
	title := colorString(colorCyan, getTitle(response))
	if response.ContentLength < 0 {
		response.ContentLength = 0
	}
	var cl string
	if Csv {
		cl = colorString(colorMajenta, fmt.Sprintf("%v", response.ContentLength))
	} else {
		cl = colorString(colorMajenta, fmt.Sprintf("Content-Length:%v", response.ContentLength))
	}
	result := fmt.Sprintf(formatString, url, statusCode, cl, title)
	if !strings.Contains(o.StatusCodeBlacklist, fmt.Sprint(response.StatusCode)) {
		fmt.Println(result)
		if file != nil {
			file.WriteString(result + "\n")
		}
	}
}

func checkStatusCodeBlacklist() bool {
	re := regexp.MustCompile("^[1-5][0-9][0-9]$")
	for _, element := range strings.Split(o.StatusCodeBlacklist, ",") {
		if !re.MatchString(element) {
			return false
		}
	}
	return true
}

func showHelper() {
	helper := []string{
		"hike by zblurx",
		"",
		"Usage: hike [flags] url",
		"",
		"Request a URL by spliting the path like:",
		"curl https://path/to/a/big/rce",
		"curl https://path/to/a/big",
		"curl https://path/to/a",
		"curl https://path/to",
		"curl https://path",
		"",
		"Can do a full permutation too (lot of requests)",
		"",
		" -u, --url <url>\t\t\tSpecify URL",
		" -H, --header <header>\t\t\tSpecify header. Can be used multiple times",
		" -c, --cookies <cookies>\t\tSpecify cookies",
		" -x, --proxy <proxy>\t\t\tSpecify proxy",
		" -p, --permute\t\t\t\tPermute Url path combination",
		" -k, --insecure\t\t\t\tAllow insecure server connections when using SSL",
		" -t, --threads <int>\t\t\tNumber of thread. Default 10",
		" -b, --status-code-blacklist <list>\tComme separated list of status code not to output",
		" -csv\t\t\t\t\tCSV output comma separated. Plus no color automatically",
		"     --no-color\t\t\t\tNo color output, boring",
		"     --no-request\t\t\tJust output urls",
	}

	fmt.Println(strings.Join(helper, "\n"))
}

func init() {
	flag.Usage = func() {
		showHelper()
	}
}

func main() {

	flag.StringVar(&o.Cookies, "cookies", "", "")
	flag.StringVar(&o.Cookies, "c", "", "")

	flag.StringVar(&o.Proxy, "proxy", "", "")
	flag.StringVar(&o.Proxy, "x", "", "")

	flag.StringVar(&o.Url, "url", "", "")
	flag.StringVar(&o.Url, "u", "", "")

	flag.Var(&o.Headers, "header", "")
	flag.Var(&o.Headers, "H", "")

	flag.BoolVar(&o.NoRequest, "no-request", false, "")

	flag.BoolVar(&o.NoColor, "no-color", false, "")

	flag.BoolVar(&o.Csv, "csv", false, "")

	flag.BoolVar(&o.Permute, "permute", false, "")
	flag.BoolVar(&o.Permute, "p", false, "")

	flag.BoolVar(&o.NotCheckCert, "insecure", false, "")
	flag.BoolVar(&o.NotCheckCert, "k", false, "")

	flag.StringVar(&o.StatusCodeBlacklist, "status-code-blacklist", "", "")
	flag.StringVar(&o.StatusCodeBlacklist, "b", "", "")

	flag.StringVar(&o.OutputFile, "output-file", "", "")
	flag.StringVar(&o.OutputFile, "o", "", "")

	flag.Parse()

	if o.Url == "" {
		showHelper()
		os.Exit(1)
	}

	var wg sync.WaitGroup
	var pg sync.WaitGroup

	requests := make(chan request)
	responses := make(chan response)

	if o.StatusCodeBlacklist != "" && !checkStatusCodeBlacklist() {
		fmt.Println("Status Code Blacklist is not correct")
		os.Exit(1)
	}

	wg.Add(1)
	go func() {
		for req := range requests {
			responses <- response{
				RequestURL: req.RequestURL,
				Response:   getResponseFromURL(req),
			}
		}
		wg.Done()
	}()

	pg.Add(1)
	go func() {
		var output_file *os.File
		if o.OutputFile != "" {
			var err error
			output_file, err = os.Create(o.OutputFile)
			if isError(err) {
				os.Exit(1)
			}

			defer output_file.Close()
		}
		formatString := "%s [%s] [%s] [%s]"
		if o.Csv {
			formatString = "%s,%s,%s,%s"
		}
		for resp := range responses {
			printResponse(resp.Response, resp.RequestURL, output_file, formatString, o.NoColor, o.Csv)
		}
		pg.Done()
	}()

	host := setHostHeaderIfExists()

	httpClient := prepareClient()

	baseUrl, path := splitUrl(o.Url)
	var urls []string
	if o.Permute {
		urls = getPermutation(baseUrl, path)
	} else {
		urls = hikePath(baseUrl, path)
	}
	if o.NoRequest {
		for _, e := range urls {
			fmt.Println(e)
		}
	} else {
		for _, e := range urls {
			requests <- request{
				RequestURL: e,
				Client:     httpClient,
				Request:    prepareRequest(e, host),
			}
		}
	}

	close(requests)
	wg.Wait()
	close(responses)
	pg.Wait()
}
