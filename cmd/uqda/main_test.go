package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestPrintUsageGuidesCommonTasks(t *testing.T) {
	oldCommandLine := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet("uqda", flag.ContinueOnError)
	t.Cleanup(func() { flag.CommandLine = oldCommandLine })
	flag.CommandLine.Bool("address", false, "test option")

	var output bytes.Buffer
	printUsage(&output, "uqda")
	usage := output.String()

	for _, expected := range []string{
		"uqda [options]",
		"uqdactl doctor",
		"uqdactl getSelf",
		"uqdactl getPeers",
		"uqda -autoconf",
		"uqda -genconf > uqda.conf",
		"uqda -useconffile uqda.conf -address",
		"-address, -subnet, -publickey, -exportkey and -normaliseconf require",
		"-user is a daemon/service option",
		"-nameserver belong to uqda-stack",
		"Options:",
		"-address",
	} {
		if !strings.Contains(usage, expected) {
			t.Errorf("usage does not contain %q\nfull output:\n%s", expected, usage)
		}
	}
}
