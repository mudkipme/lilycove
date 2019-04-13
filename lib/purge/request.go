package purge

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type requestOptions struct {
	method  string
	url     string
	headers map[string]string
}

func (p *Purger) handlePurge(item purgeItem) {
	ros := []requestOptions{}
	u, err := url.Parse(item.Host + item.URL)
	if err != nil {
		return
	}
	queries := strings.Split(u.RawQuery, "&")
	lastQuery := ""
	if len(queries) > 0 {
		lastQuery = queries[len(queries)-1]
	}
	pathComponents := strings.Split(u.RawPath, "/")
	firstPath := ""
	if len(pathComponents) > 1 {
		firstPath = pathComponents[1]
	}

	for _, entry := range p.config.Entries {
		if entry.Host != item.Host {
			continue
		}
		variants := make(map[string]bool)
		variants[""] = true
		for _, variant := range entry.Variants {
			variants[variant] = true
		}

		for _, uri := range entry.URIs {
			for variant := range variants {
				if variant != "" && ((lastQuery != "" && variants[lastQuery]) ||
					(firstPath != "" && variants[firstPath])) {
					continue
				}
				ros = append(ros, requestOptions{
					method:  entry.Method,
					url:     strings.ReplaceAll(strings.ReplaceAll(uri, "#url#", item.URL), "#variants#", variant),
					headers: entry.Headers,
				})

				if strings.Contains(item.URL, "/wiki/") && variant != "" {
					ros = append(ros, requestOptions{
						method:  entry.Method,
						url:     strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(uri, "#url#", item.URL), "#variants#", ""), "/wiki/", "/"+variant+"/"),
						headers: entry.Headers,
					})
				}
			}
		}
	}

	ros = uniq(ros)
	ch := make(chan bool)
	for _, ro := range ros {
		go p.doRequest(ro.method, ro.url, ro.headers, ch)
	}
	for range ros {
		<-ch
	}
}

func uniq(input []requestOptions) (res []requestOptions) {
	res = make([]requestOptions, 0, len(input))
	seen := make(map[string]bool)
	for _, val := range input {
		if _, ok := seen[val.url]; !ok {
			seen[val.url] = true
			res = append(res, val)
		}
	}
	return
}

func (p *Purger) doRequest(method, url string, headers map[string]string, ch chan bool) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Printf("[Purger] Error sending purge request. %v %v\n", url, err)
		ch <- false
		return
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}
	response, err := p.client.Do(req)
	if err != nil {
		fmt.Printf("[Purger] Error sending purge request. %v %v\n", url, err)
		ch <- false
	} else {
		response.Body.Close()
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			fmt.Printf("[Purger] Purge success. %v\n", url)
		} else if response.StatusCode != http.StatusNotFound && response.StatusCode >= 300 {
			fmt.Printf("[Purger] Error sending purge request. %v %d\n", url, response.StatusCode)
		}
		ch <- true
	}
}
