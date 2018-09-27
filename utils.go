package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func difference(a, b []string) (diff []string) {
	m := make(map[string]bool)
	for _, item := range b {
		m[item] = true
	}
	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return diff
}

func getUserHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func expandUser(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}
	home, err := getUserHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}

func checkFilePermissions(filename string, perm int) error {
	file, err := os.OpenFile(filename, perm, 0644)
	defer file.Close()
	if err != nil {
		if os.IsPermission(err) {
			return err
		}
	}
	return nil
}

func writeToFileStrings(filename string, strings []string, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, text := range strings {
		if _, err = f.WriteString(text + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func getTelegramURL(bot string, cid string, links []string) string {
	baseURL := fmt.Sprintf(telegramLink, bot, cid)
	message := url.PathEscape(strings.Join(links, "\n\n"))
	fullURL := fmt.Sprintf(baseURL+"&text=%s", message)
	return fullURL
}

func telegramMessage(bot string, cid string, links []string, client *http.Client) error {
	telegramURL := getTelegramURL(bot, cid, links)
	return makeRequest(telegramURL, client)
}
