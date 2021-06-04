// check-gosnmp-cpu plugin for Icinga2 compatible systems
// Copyright 2021 by Marko Punnar <marko[AT]aretaja.org>
//
// Icinga2 plugin designed to check CPU load.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>
package main

import (
	"flag"
	"fmt"
	"math"
	"net"
	"os"

    "github.com/aretaja/icingahelper"
    "github.com/aretaja/snmphelper"
)

// Version of release
const Version = "0.0.1"

// Struct for cpu data gathered using HOST-RESOURCES-MIB
type cpuLoad struct {
	cpuCnt int64
	load   int64
}

// .iso.org.dod.internet.mgmt.mib-2.host.hrDevice.hrProcessorTable.hrProcessorEntry.hrProcessorLoad
const procLoad = ".1.3.6.1.2.1.25.3.3.1.2"

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
	var warn = flag.String("w", "85", "[warning level] (%)")
	var crit = flag.String("c", "95", "[critical level] (%)")
	var ver = flag.Bool("v", false, "Using this parameter will display the version number and exit")

	flag.Parse()

	// Initialize new check object
	check := icingahelper.NewCheck("CPU")

	// Show version
	if *ver {
		fmt.Println("Plugin version " + Version)
		os.Exit(check.RetVal())
	}

	// Exit if no host submitted
	if net.ParseIP(*host) == nil {
		fmt.Println("Valid host ip is required")
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
	err := session.New()
	if err != nil {
		fmt.Printf("SNMP error: %v\n", err)
		os.Exit(check.RetVal())
	}

	// Do SNMP query
	err = session.Walk(procLoad, true, true)
	if err != nil {
		fmt.Printf("SNMP error: %v\n", err)
		os.Exit(check.RetVal())
	}

	// DEBUG
	// fmt.Printf("%# v\n", pretty.Formatter(res))

	cpuData, err := calcCPUData(session.Result)
	if err != nil {
		fmt.Printf("CPU data error: %v\n", session.Result)
		os.Exit(check.RetVal())
	}

	// DEBUG
	// fmt.Printf("%# v\n", pretty.Formatter(cpuData))

	level, err := check.AlarmLevel(int64(cpuData.load), *warn, *crit)
	if err != nil {
		fmt.Printf("Alarm level error: %v\n", err)
		os.Exit(check.RetVal())
	}

	check.AddPerfData("'cpu usage'", cpuData.load, "%", 0, 100, *warn, *crit)
	check.AddMsg(level, fmt.Sprintf("%d CPUs, load %d%%", cpuData.cpuCnt, cpuData.load), "")

	// DEBUG
	// fmt.Printf("%# v\n", pretty.Formatter(check))

	fmt.Print(check.FinalMsg())
	os.Exit(check.RetVal())
}

// Returns cpu usage data
func calcCPUData(data snmphelper.SnmpOut) (*cpuLoad, error) {
	out := &cpuLoad{}
	var loads []int64

	for _, d := range data {
		loads = append(loads, d.Integer)
	}

	cnt := int64(len(loads))
	if cnt == 0 {
		return out, fmt.Errorf("CPU count 0 or unknown")
	}

	var loadSum int64 = 0
	for _, v := range loads {
		loadSum += v
	}

	var loadAvg float64 = float64(loadSum / cnt)
	var load int64 = int64(math.Round(loadAvg))

	out.cpuCnt = cnt
	out.load = load

	return out, nil
}
