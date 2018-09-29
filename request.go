package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"runtime"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

func readBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func prepareClient() (*http.Client, error) {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("Too many redirects")
			}
			if len(via) == 0 {
				return nil
			}
			for attr, val := range via[0].Header {
				if _, ok := req.Header[attr]; !ok {
					req.Header[attr] = val
				}
			}
			return nil
		},
	}
	return client, nil
}

func parsePageUrls(client *http.Client, in <-chan strResult, fout chan chan strResult, wg *sync.WaitGroup) {
	out := make(chan strResult)
	go func() {
		defer wg.Done()
		defer close(out)
		fout <- out
		for res := range in {
			if res.error != nil {
				out <- res
				continue
			}
			req, err := http.NewRequest("GET", res.result, nil)
			if err != nil {
				out <- strResult{error: err}
				continue
			}
			for key, values := range agentHeaders {
				for _, header := range values {
					req.Header.Add(key, header)
				}
			}
			resp, err := client.Do(req)
			if err != nil {
				out <- strResult{error: err}
				continue
			}
			body, err := readBody(resp)
			if err != nil {
				out <- strResult{error: err}
				continue
			}
			if resp.StatusCode != 200 {
				out <- strResult{error: fmt.Errorf("Error. Response status code: %d\n%s", resp.StatusCode, body)}
				continue
			}
			links, err := getPageLinks(body)
			if err != nil {
				out <- strResult{error: err}
				continue
			}
			for _, link := range *links {
				out <- strResult{result: link}
			}
		}
	}()
}

func parseMainPage(url string, client *http.Client) <-chan strResult {
	out := make(chan strResult)

	go func() {
		defer close(out)
		out <- strResult{result: url}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			out <- strResult{error: err}
			return
		}
		for key, values := range agentHeaders {
			for _, header := range values {
				req.Header.Add(key, header)
			}
		}
		resp, err := client.Do(req)
		if err != nil {
			out <- strResult{error: err}
			return
		}
		body, err := readBody(resp)
		if err != nil {
			out <- strResult{error: err}
			return
		}
		if resp.StatusCode != 200 {
			out <- strResult{error: fmt.Errorf("Error. Response status code: %d\n%s", resp.StatusCode, body)}
			return
		}
		links, err := getPages(body)
		if err != nil {
			out <- strResult{error: err}
			return
		}
		for _, link := range *links {
			out <- strResult{result: link}
		}
	}()
	return out
}

func parsePages(in <-chan strResult, client *http.Client) <-chan chan strResult {
	out := make(chan chan strResult)
	var wg sync.WaitGroup
	cpu := runtime.NumCPU()
	wg.Add(cpu)
	for i := 0; i < cpu; i++ {
		go parsePageUrls(client, in, out, &wg)
	}
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

func makeRequest(url string, client *http.Client) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	for key, values := range agentHeaders {
		for _, header := range values {
			req.Header.Add(key, header)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body, err := readBody(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error. Response status code: %d\n%s", resp.StatusCode, body)
	}
	return nil
}
