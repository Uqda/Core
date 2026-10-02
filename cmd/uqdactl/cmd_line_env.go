package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/hjson/hjson-go/v4"
	"golang.org/x/text/encoding/unicode"

	"github.com/Uqda/Core/internal/cli"
	"github.com/Uqda/Core/src/config"
	"github.com/Uqda/Core/src/version"
)

type cmdLineEnv struct {
	args                 []string
	endpoint, server     string
	injson, borders, ver bool
}

func newCmdLineEnv() cmdLineEnv {
	var cmdLineEnv cmdLineEnv
	cmdLineEnv.endpoint = config.GetDefaults().DefaultAdminListen
	return cmdLineEnv
}

func (cmdLineEnv *cmdLineEnv) parseFlagsAndArgs(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("uqdactl", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.Usage = func() {
		cli.Help(output, "uqdactl", version.DisplayName())
		fmt.Fprintln(output, "  Options (before or after the command):")
		flags.PrintDefaults()
	}
	server := flags.String("endpoint", cmdLineEnv.endpoint, "Local admin socket endpoint")
	injson := flags.Bool("json", false, "Machine-readable JSON output")
	borders := flags.Bool("borders", true, "Borders on advanced tables")
	ver := flags.Bool("version", false, "Print the installed release")
	var options, positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}
		options = append(options, arg)
		name := strings.SplitN(strings.TrimLeft(arg, "-"), "=", 2)[0]
		if name == "endpoint" && !strings.Contains(arg, "=") {
			if i+1 == len(args) || strings.HasPrefix(args[i+1], "-") {
				return fmt.Errorf("-endpoint requires a value")
			}
			i++
			options = append(options, args[i])
		}
	}
	if err := flags.Parse(options); err != nil {
		return err
	}
	if *ver && len(positional) != 0 {
		return fmt.Errorf("-version cannot be combined with a command")
	}
	if len(positional) == 0 {
		positional = []string{"status"}
	}
	if strings.EqualFold(positional[0], "help") {
		if len(positional) == 2 && cli.HelpTopic(output, "uqdactl", version.DisplayName(), positional[1]) {
			return flag.ErrHelp
		}
		if len(positional) != 1 {
			return fmt.Errorf("use help or help COMMAND with a known command")
		}
		flags.Usage()
		return flag.ErrHelp
	}
	aliases := map[string]string{"peers": "getpeers", "info": "getself", "commands": "list"}
	if canonical, ok := aliases[strings.ToLower(positional[0])]; ok {
		positional[0] = canonical
	}
	name := strings.ToLower(positional[0])
	if (isDoctorCommand(name) || name == "version") && len(positional) != 1 {
		return fmt.Errorf("%s takes no arguments", name)
	}
	if name != "test" {
		seen := make(map[string]bool)
		for i, arg := range positional[1:] {
			key, _, ok := strings.Cut(arg, "=")
			if !ok || key == "" || seen[key] {
				return fmt.Errorf("invalid or duplicate argument at position %d; use name=value", i+1)
			}
			seen[key] = true
		}
	}
	cmdLineEnv.args = positional
	cmdLineEnv.server = *server
	cmdLineEnv.injson = *injson
	cmdLineEnv.borders = *borders
	cmdLineEnv.ver = *ver
	return nil
}

func (cmdLineEnv *cmdLineEnv) setEndpoint(logger *log.Logger) error {
	if cmdLineEnv.server == cmdLineEnv.endpoint {
		if cfg, err := os.ReadFile(config.GetDefaults().DefaultConfigFile); err == nil {
			if len(cfg) >= 2 && (bytes.Equal(cfg[0:2], []byte{0xFF, 0xFE}) ||
				bytes.Equal(cfg[0:2], []byte{0xFE, 0xFF})) {
				utf := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
				decoder := utf.NewDecoder()
				cfg, err = decoder.Bytes(cfg)
				if err != nil {
					return fmt.Errorf("decode configuration file %q: %w", config.GetDefaults().DefaultConfigFile, err)
				}
			}
			var dat map[string]interface{}
			if err := hjson.Unmarshal(cfg, &dat); err != nil {
				return fmt.Errorf("parse configuration file %q: %w", config.GetDefaults().DefaultConfigFile, err)
			}
			if ep, ok := dat["AdminListen"].(string); ok && (ep != "none" && ep != "") {
				cmdLineEnv.endpoint = ep
				logger.Println("Found platform default config file", config.GetDefaults().DefaultConfigFile)
				logger.Println("Using endpoint", cmdLineEnv.endpoint, "from AdminListen")
			} else {
				logger.Println("Configuration file doesn't contain appropriate AdminListen option")
				logger.Println("Falling back to platform default", config.GetDefaults().DefaultAdminListen)
			}
		} else {
			logger.Println("Can't open config file from default location", config.GetDefaults().DefaultConfigFile)
			logger.Println("Falling back to platform default", config.GetDefaults().DefaultAdminListen)
		}
	} else {
		cmdLineEnv.endpoint = cmdLineEnv.server
		logger.Println("Using endpoint", cmdLineEnv.endpoint, "from command line")
	}
	return nil
}
