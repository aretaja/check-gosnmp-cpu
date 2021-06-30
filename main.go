// Copyright 2021 by Marko Punnar <marko[AT]aretaja.org>
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

// check-gosnmp-cpu is CPU load plugin for Icinga2 compatible systems
package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/aretaja/check-gosnmp-cpu/cpu"
	"github.com/aretaja/icingahelper"
	"github.com/aretaja/snmphelper"
)

// Version of release
const Version = "1.0.1"

func main() {
	// Parse cli arguments
	var host = flag.String("H", "", "<host ip>")
	var snmpVer = flag.Int("V", 2, "[snmp version] (1|2|3)")
	var snmpUser = flag.String("u", "public", "[username|community]")
	var snmpProt = flag.String("a", "MD5", "[authentication protocol] (NoAuth|MD5|SHA)5")
	var snmpPass = flag.String("A", "", "[authentication protocol pass phrase]")
	var snmpSlevel = flag.String("l", "authPriv", "[security level] (noAuthNoPriv|authNoPriv|authPriv)")
	var snmpPrivProt = flag.String("x", "DES", "[privacy protocol] (NoPriv|DES|AES|AES192|AES256|AES192C|AES256C)")
	var snmpPrivPass = flag.String("X", "", "[privacy protocol pass phrase]")
	var warn = flag.String("w", "85", "[warning level]. It depends of check type.\n"+
		"\thost - % of average cpu utilization of all cores\n"+
		"\tsystat - % of cpu utilization\n"+
		"\tloadavg - summary % of 1 min. load on all cores fe. 100 means that load for last min is 4 when device has 4 cpu cores\n"+
		"\t\t5 and 15 minute levels will be calculated from this value by decreasing them by 5 and 10 accordingly\n"+
		"\tjnx - % of cpu utilization\n"+
		"\tcisco - overall cpu busy % in the last 1 minute period\n"+
		"\t\t5 minute level will be calculated from this value by decreasing value by 5\n"+
		"\trcsw - % of cpu utilization\n"+
		"\tmoxasw - overall cpu busy % in the last 5 sec period\n"+
		"\t\t30 sec and 5 minute levels will be calculated from this value by decreasing value by 5 and 10 accordingly",
	)
	var crit = flag.String("c", "95", "[critical level]. Look at warning level explanation")
	var ctype = flag.String("t", "", "<check type>\n"+
		"\thost - uses hostmib\n"+
		"\tsysstats - uses UCD-SNMP-MIB systemStats\n"+
		"\tloadavg - uses UCD-SNMP-MIB laTable\n"+
		"\tjnx - uses jnxOperatingTable\n"+
		"\tcisco - uses ciscoProcessMIB\n"+
		"\trcsw - uses rcDeviceStsCpuUsagePercent\n"+
		"\tmoxasw - uses moxa MIB",
	)
	var dbg = flag.Bool("d", false, "Using this parameter will print out debug info")
	var ver = flag.Bool("v", false, "Using this parameter will display the version number and exit")

	flag.Parse()

	// Initialize new check object
	check := icingahelper.NewCheck("CPU")

	// Show version
	if *ver {
		fmt.Println("plugin version " + Version)
		os.Exit(check.RetVal())
	}

	// Exit if no host submitted
	if net.ParseIP(*host) == nil {
		fmt.Println("valid host ip is required")
		os.Exit(check.RetVal())
	}

	// Exit if no type submitted
	if *ctype == "" {
		fmt.Println("check type required")
		os.Exit(check.RetVal())
	}

	// Session variables
	session := snmphelper.Session{
		Host:     *host,
		Ver:      *snmpVer,
		User:     *snmpUser,
		Prot:     *snmpProt,
		Pass:     *snmpPass,
		Slevel:   *snmpSlevel,
		PrivProt: *snmpPrivProt,
		PrivPass: *snmpPrivPass,
	}

	// Initialize session
	sess, err := session.New()
	if err != nil {
		fmt.Printf("snmp error: %v\n", err)
		os.Exit(check.RetVal())
	}

	// Get CPU load
	load := cpu.Load{
		Check: check,
		Sess:  sess,
		Warn:  *warn,
		Crit:  *crit,
		Ctype: *ctype,
		Debug: *dbg,
	}

	err = load.Get()
	if err != nil {
		fmt.Println(err)
		os.Exit(check.RetVal())
	}

	fmt.Print(check.FinalMsg())
	os.Exit(check.RetVal())
}
