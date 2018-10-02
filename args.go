package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/alyu/configparser"
)

func prepareArgs() neagentArgs {
	databasePtr := flag.String("db", "", "Database file")
	filePtr := flag.String("f", "", "Output file")
	linkPtr := flag.String("l", "", "Neagent link")
	configPtr := flag.String("c", "", "Config file")
	botPtr := flag.String("bot", "", "Telegram bot token")
	cidPtr := flag.String("cid", "", "Telegram chat id")
	timeoutPtr := flag.Int("t", defaultTimeout, "Requests timeout")
	verbosePtr := flag.Bool("v", false, "Verbose mode")
	daemonPtr := flag.Bool("d", false, "Daemon mode")
	stdoutPtr := flag.Bool("s", false, "Print results to stdout")
	telPtr := flag.Bool("tel", false, "Send results to telegram")
	flag.Parse()
	return neagentArgs{
		*databasePtr,
		*filePtr,
		*linkPtr,
		*configPtr,
		*botPtr,
		*cidPtr,
		*timeoutPtr,
		*verbosePtr,
		*daemonPtr,
		*stdoutPtr,
		*telPtr,
	}
}

func setStringOption(args *neagentArgs, name string, options map[string]string) {
	value, exists := options[strings.ToLower(name)]
	if exists && len(value) > 0 {
		reflect.ValueOf(args).Elem().FieldByName(name).SetString(strings.Trim(value, "\"'`"))
	}
}

func setIntOption(args *neagentArgs, name string, options map[string]string) error {
	value, exists := options[strings.ToLower(name)]
	if exists {
		res, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		reflect.ValueOf(args).Elem().FieldByName(name).SetInt(res)
	}
	return nil
}

func setBoolOption(args *neagentArgs, name string, options map[string]string) error {
	value, exists := options[strings.ToLower(name)]
	if exists {
		res, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		reflect.ValueOf(args).Elem().FieldByName(name).SetBool(res)
	}
	return nil
}

func readConfig(args *neagentArgs, filename string) error {
	config, err := configparser.Read(filename)
	if err != nil {
		return err
	}
	sections, err := config.AllSections()
	if err != nil {
		return err
	}
	options := sections[0].Options()
	for _, flag := range []string{"Database", "File", "Link", "Bot", "Cid"} {
		setStringOption(args, flag, options)
	}
	for _, flag := range []string{"Timeout"} {
		setIntOption(args, flag, options)
	}
	for _, flag := range []string{"Verbose", "Daemon", "Stdout", "Telegram"} {
		err := setBoolOption(args, flag, options)
		if err != nil {
			return err
		}
	}
	args.Config = filename
	return nil
}

func readConfigs(filename string) (*neagentArgs, error) {
	home, err := getUserHome()
	if err != nil {
		return nil, err
	}
	homeConfig, _ := filepath.Abs(filepath.Join(home, defaultConfigFile))
	currentConfig, _ := filepath.Abs(defaultConfigFile)
	configs := []string{homeConfig, currentConfig}
	if len(filename) > 0 {
		argsConfig, _ := filepath.Abs(filename)
		configs = append(configs, argsConfig)
	}
	args := &neagentArgs{"", "", "", "", "", "", 0, false, false, false, false}
	for _, name := range configs {
		readConfig(args, name)
	}
	return args, nil
}

func checkConfig(args *neagentArgs) error {
	if len(args.Config) > 0 {
		filename, _ := filepath.Abs(args.Config)
		if stat, err := os.Stat(filename); os.IsNotExist(err) || stat.IsDir() {
			return fmt.Errorf("There is no the config file or file is a directory: %s", filename)
		}
	}
	return nil
}

func checkArgs(args *neagentArgs) error {
	if args.Telegram {
		if len(args.Bot) == 0 {
			return fmt.Errorf("No bot argument has been supplied")
		}
		if len(args.Cid) == 0 {
			return fmt.Errorf("No cid argument has been supplied")
		}
	}
	if len(args.Database) == 0 {
		return fmt.Errorf("No database argument has been supplied")
	}
	if len(args.Link) == 0 {
		return fmt.Errorf("No link argument has been supplied")
	}
	if args.Timeout == 0 {
		return fmt.Errorf("No timeout argument has been supplied")
	}
	err := checkFilePermissions(args.Database, os.O_WRONLY|os.O_CREATE)
	if err != nil {
		return err
	}
	if len(args.File) > 0 {
		err = checkFilePermissions(args.File, os.O_WRONLY|os.O_CREATE)
		if err != nil {
			return err
		}
	}
	return nil
}

func ajustArgs(args *neagentArgs) error {
	db, err := expandUser(args.Database)
	if err != nil {
		return err
	}
	args.Database = db
	if len(args.File) > 0 {
		file, err := expandUser(args.File)
		if err != nil {
			return err
		}
		args.File = file
	}
	return nil
}

func mergeConfigs(args *neagentArgs, configArgs *neagentArgs) *neagentArgs {
	if len(args.Database) == 0 && len(configArgs.Database) > 0 {
		args.Database = configArgs.Database
	}
	if len(args.File) == 0 && len(configArgs.File) > 0 {
		args.File = configArgs.File
	}
	if len(args.Link) == 0 && len(configArgs.Link) > 0 {
		args.Link = configArgs.Link
	}
	if len(args.Cid) == 0 && len(configArgs.Cid) > 0 {
		args.Cid = configArgs.Cid
	}
	if len(args.Bot) == 0 && len(configArgs.Bot) > 0 {
		args.Bot = configArgs.Bot
	}
	if args.Timeout == 0 && args.Timeout > 0 {
		args.Timeout = configArgs.Timeout
	}
	if configArgs.Stdout {
		args.Stdout = configArgs.Stdout
	}
	if configArgs.Verbose {
		args.Verbose = configArgs.Verbose
	}
	if configArgs.Daemon {
		args.Daemon = configArgs.Daemon
	}
	if configArgs.Telegram {
		args.Telegram = configArgs.Telegram
	}
	return args
}
