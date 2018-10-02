package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

func failIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func prepareParams() (*neagentArgs, error) {
	args := prepareArgs()
	err := checkConfig(&args)
	if err != nil {
		return nil, err
	}
	configArgs, err := readConfigs(args.Config)
	if err != nil {
		return nil, err
	}
	args = *mergeConfigs(&args, configArgs)
	err = ajustArgs(&args)
	if err != nil {
		return nil, err
	}
	err = checkArgs(&args)
	if err != nil {
		return nil, err
	}
	return &args, nil
}

func iteration(args *neagentArgs, db *sql.DB) {
	if args.Verbose {
		log.Println("Starting processing...")
	}
	client, err := prepareClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	ctx, done := context.WithCancel(context.Background())
	defer done()
	cn := parsePages(parseMainPage(args, client), client)
	var links []string
	for res := range orDone(ctx, bridge(ctx, cn)) {
		if res.error != nil {
			fmt.Println(res.error)
			continue
		}
		links = append(links, res.result)
	}
	newLinks, err := getNewLinks(args.Link, links, db)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = storeLinks(args.Link, links, db)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(newLinks) == 0 {
		return
	}
	if args.Stdout {
		for _, link := range newLinks {
			fmt.Println(link)
		}
	}
	if len(args.File) > 0 {
		err = writeToFileStrings(args.File, newLinks, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error on file writing: %v\n", err)
		}
	}
	if args.Telegram {
		if args.Verbose {
			log.Println("Sending telegram message...")
		}
		err = telegramMessage(args.Bot, args.Cid, newLinks, client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error on telegram messaging: %v\n", err)
		}
	}
}

func startProcessing(args *neagentArgs, database *sql.DB) {
	iteration(args, database)
	for range time.NewTicker(time.Duration(args.Timeout) * time.Second).C {
		iteration(args, database)
	}
}

func main() {
	args, err := prepareParams()
	failIfError(err)
	if args.Verbose {
		log.Println("Creating database...")
	}
	database, err := createDataBase(args.Database)
	failIfError(err)
	if args.Daemon {
		startDaemon(startProcessing, args, database)
	} else {
		startProcessing(args, database)
	}
}
