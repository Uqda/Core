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
		"Examples for every option:",
		"-address         uqda -useconffile uqda.conf -address",
		"-autoconf        uqda -autoconf",
		"-exportkey       uqda -useconffile uqda.conf -exportkey",
		"-genconf         uqda -genconf > uqda.conf",
		"-json            uqda -genconf -json > uqda.json",
		"-loglevel        uqda -autoconf -loglevel debug",
		"-logto           uqda -autoconf -logto uqda.log",
		"-normaliseconf   uqda -useconffile uqda.conf -normaliseconf",
		"-notifyfd        Service-manager integration only",
		"-publickey       uqda -useconffile uqda.conf -publickey",
		"-subnet          uqda -useconffile uqda.conf -subnet",
		"-useconf         cat uqda.conf | uqda -useconf",
		"-useconffile     uqda -useconffile uqda.conf",
		"-user            uqda -useconffile uqda.conf -user USER:GROUP",
		"-version         uqda -version",
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
