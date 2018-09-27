package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS links (
    id INTEGER PRIMARY KEY,
    base_link TEXT,
    link TEXT,
    added INTEGER
);
CREATE UNIQUE INDEX IF NOT EXISTS u_links ON links(base_link, link);
`

func prepareSelectQuery(baseLink string, links []string) string {
	var quotedLinks []string
	for _, link := range links {
		quotedLinks = append(quotedLinks, fmt.Sprintf("'%s'", link))
	}
	linksReq := fmt.Sprintf("(%s)", strings.Join(quotedLinks, ","))
	return fmt.Sprintf(
		"SELECT link FROM links WHERE base_link = '%s' AND link IN %s",
		baseLink,
		linksReq,
	)
}

func prepareInsertQuery(baseLink string, links []string) string {
	var queryValues []string
	for _, link := range links {
		queryValues = append(queryValues, fmt.Sprintf(
			"('%s', '%s', CURRENT_TIMESTAMP)",
			baseLink,
			link,
		))
	}
	return fmt.Sprintf(
		"INSERT INTO links(base_link, link, added) VALUES %s",
		strings.Join(queryValues, ", "),
	)
}

func createDataBase(filename string) (*sql.DB, error) {
	database, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	statement, err := database.Prepare(createTableSQL)
	if err != nil {
		return nil, err
	}
	statement.Exec()
	return database, nil
}

func getNewLinks(baseLink string, links []string, database *sql.DB) ([]string, error) {
	query := prepareSelectQuery(baseLink, links)
	rows, err := database.Query(query)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var link string
	var existingLinks, newLinks []string
	for rows.Next() {
		rows.Scan(&link)
		existingLinks = append(existingLinks, link)
	}
	newLinks = difference(links, existingLinks)
	return newLinks, nil
}

func storeLinks(baseLink string, links []string, database *sql.DB) error {
	if len(links) > 0 {
		query := prepareInsertQuery(baseLink, links)
		statement, err := database.Prepare(query)
		defer statement.Close()
		if err != nil {
			return err
		}
		statement.Exec()
	}
	return nil
}
