package main

import (
	"database/sql"
	"log"

	"github.com/sevlyar/go-daemon"
)

type fn func(args *neagentArgs, database *sql.DB)

func startDaemon(fnc fn, args *neagentArgs, database *sql.DB) {
	cntxt := &daemon.Context{
		PidFileName: "neagent.pid",
		PidFilePerm: 0644,
		LogFileName: "neagent.log",
		LogFilePerm: 0640,
		WorkDir:     daemonWorkDir,
		Umask:       027,
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("daemon started")

	fnc(args, database)
}
