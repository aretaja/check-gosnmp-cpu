// package cpu implements CPU load measurement
package cpu

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/aretaja/icingahelper"
	"github.com/aretaja/snmphelper"
	"github.com/kr/pretty"
)

// Struct for cpu data gathered using HOST-RESOURCES-MIB
type Load struct {
	Check             *icingahelper.IcingaCheck
	Sess              *snmphelper.Session
	Warn, Crit, Ctype string
	Debug             bool
}

// .iso.org.dod.internet.mgmt.mib-2.system.sysObjectID
const sysObjectID = ".1.3.6.1.2.1.1.2.0"

// .iso.org.dod.internet.mgmt.mib-2.host.hrDevice.hrProcessorTable.hrProcessorEntry.hrProcessorLoad
const hrProcessorLoad = ".1.3.6.1.2.1.25.3.3.1.2"

// .iso.org.dod.internet.private.enterprises.ucdavis.systemStats.ssCpuUser
const ssCpuUser = ".1.3.6.1.4.1.2021.11.9.0"

// .iso.org.dod.internet.private.enterprises.ucdavis.systemStats.ssCpuSystem
const ssCpuSystem = ".1.3.6.1.4.1.2021.11.10.0"

// .iso.org.dod.internet.private.enterprises.ucdavis.systemStats.ssCpuIdle
const ssCpuRawIdle = ".1.3.6.1.4.1.2021.11.11.0"

// .iso.org.dod.internet.private.enterprises.ucdavis.laTable.laEntry.laLoadInt
const laLoadInt = ".1.3.6.1.4.1.2021.10.1.5"

// .iso.org.dod.internet.private.enterprises.juniperMIB.jnxMibs.jnxBoxAnatomy.jnxOperatingTable.jnxOperatingEntry.jnxOperatingDescr
const jnxOperatingDescr = ".1.3.6.1.4.1.2636.3.1.13.1.5"

// .iso.org.dod.internet.private.enterprises.juniperMIB.jnxMibs.jnxBoxAnatomy.jnxOperatingTable.jnxOperatingEntry.jnxOperatingCPU
const jnxOperatingCPU = ".1.3.6.1.4.1.2636.3.1.13.1.8"

// .iso.org.dod.internet.private.enterprises.juniperMIB.jnxMibs.jnxBoxAnatomy.jnxOperatingTable.jnxOperatingEntry.jnxOperating1MinLoadAvg
const jnxOperating1MinLoadAvg = ".1.3.6.1.4.1.2636.3.1.13.1.20"

// .iso.org.dod.internet.private.enterprises.juniperMIB.jnxMibs.jnxBoxAnatomy.jnxOperatingTable.jnxOperatingEntry.jnxOperating5MinLoadAvg
const jnxOperating5MinLoadAvg = ".1.3.6.1.4.1.2636.3.1.13.1.21"

// .iso.org.dod.internet.private.enterprises.cisco.ciscoMgmt.ciscoProcessMIB.ciscoProcessMIBObjects.cpmCPU.cpmCPUTotalTable.cpmCPUTotalEntry.cpmCPUTotalPhysicalIndex
const cpmCPUTotalPhysicalIndex = ".1.3.6.1.4.1.9.9.109.1.1.1.1.2"

// .iso.org.dod.internet.private.enterprises.cisco.ciscoMgmt.ciscoProcessMIB.ciscoProcessMIBObjects.cpmCPU.cpmCPUTotalTable.cpmCPUTotalEntry.cpmCPUTotal1minRev
const cpmCPUTotal1minRev = ".1.3.6.1.4.1.9.9.109.1.1.1.1.7"

// .iso.org.dod.internet.private.enterprises.cisco.ciscoMgmt.ciscoProcessMIB.ciscoProcessMIBObjects.cpmCPU.cpmCPUTotalTable.cpmCPUTotalEntry.cpmCPUTotal5minRev
const cpmCPUTotal5minRev = ".1.3.6.1.4.1.9.9.109.1.1.1.1.8"

// .iso.org.dod.internet.mgmt.mib-2.entityMIB.entityMIBObjects.entityPhysical.entPhysicalTable.entPhysicalEntry.entPhysicalName
const entPhysicalName = ".1.3.6.1.2.1.47.1.1.1.1.7"

// .iso.org.dod.internet.private.enterprises.ruggedcom.ruggedcomMgmt.rcSysInfo.rcDeviceStatus.rcDeviceStsCpuUsagePercent
const rcDeviceStsCpuUsagePercent = ".1.3.6.1.4.1.15004.4.2.2.6.0"

// Do the work
func (l *Load) Get() error {
	switch l.Ctype {
	case "host":
		err := l.hostLoad()
		if err != nil {
			return err
		}
	case "sysstats":
		err := l.cpuLoad()
		if err != nil {
			return err
		}
	case "loadavg":
		err := l.sysLoad()
		if err != nil {
			return err
		}
	case "jnx":
		err := l.jnxLoad()
		if err != nil {
			return err
		}
	case "cisco":
		err := l.ciscoLoad()
		if err != nil {
			return err
		}
	case "rcsw":
		err := l.ruggedSwLoad()
		if err != nil {
			return err
		}
	case "moxasw":
		err := l.moxaSwLoad()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("no such check type")
	}

	return nil
}

// Get load data using hrProcessorLoad oid
func (l *Load) hostLoad() error {
	// Do SNMP query
	res, err := l.Sess.Walk(hrProcessorLoad, true, true)
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	cpuData, err := calcCPUData(res)
	if err != nil {
		return fmt.Errorf("cpu data error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(cpuData))
	}

	level, err := l.Check.AlarmLevel(int64(cpuData["load"]), l.Warn, l.Crit)
	if err != nil {
		return fmt.Errorf("alarm level error: %v", err)
	}

	l.Check.AddPerfData("'cpu usage'", fmt.Sprintf("%d", cpuData["load"]), "%", l.Warn, l.Crit, "0", "100")
	l.Check.AddPerfData("'cpu count'", fmt.Sprintf("%d", cpuData["cpuCnt"]), "", "", "", "", "")
	l.Check.AddPerfData("dummy", "0", "", "", "", "", "")
	l.Check.AddMsg(level, fmt.Sprintf("%d CPUs; load %d%%", cpuData["cpuCnt"], cpuData["load"]), "")

	return nil
}

// Get load data using ssCpuIdle oid
func (l *Load) cpuLoad() error {
	// Do SNMP query
	res, err := l.Sess.Get([]string{ssCpuUser, ssCpuSystem, ssCpuRawIdle})
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	d := map[string]int64{
		"used": 100 - int64(res[ssCpuRawIdle].Integer),
		"user": int64(res[ssCpuUser].Integer),
		"sys":  int64(res[ssCpuSystem].Integer),
	}

	level, err := l.Check.AlarmLevel(d["used"], l.Warn, l.Crit)
	if err != nil {
		return fmt.Errorf("alarm level error: %v", err)
	}

	l.Check.AddPerfData("cpu_prct_used", fmt.Sprintf("%d", d["used"]), "%", l.Warn, l.Crit, "0", "100")
	l.Check.AddPerfData("cpu_prct_user", fmt.Sprintf("%d", d["user"]), "%", "", "", "0", "100")
	l.Check.AddPerfData("cpu_prct_system", fmt.Sprintf("%d", d["sys"]), "%", "", "", "0", "100")
	l.Check.AddMsg(level, fmt.Sprintf("load %d%%", d["used"]), "")
	l.Check.AddMsg(level, fmt.Sprintf("user %d%%", d["user"]), "")
	l.Check.AddMsg(level, fmt.Sprintf("system %d%%", d["sys"]), "")

	return nil
}

// Get load data using laLoadInt oid
func (l *Load) sysLoad() error {
	// Get processor count
	res, err := l.Sess.Walk(hrProcessorLoad, true, true)
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	pCnt := len(res)
	if pCnt == 0 {
		return fmt.Errorf("get processor count failed: %v", err)
	}

	wPerc, err := strconv.Atoi(l.Warn)
	if err != nil {
		return fmt.Errorf("warning level must be integer: %v", err)
	}

	cPerc, err := strconv.Atoi(l.Crit)
	if err != nil {
		return fmt.Errorf("critical level must be integer: %v", err)
	}

	w1 := pCnt * wPerc
	c1 := pCnt * cPerc
	w5 := pCnt * (wPerc - 5)
	c5 := pCnt * (cPerc - 5)
	w15 := pCnt * (wPerc - 10)
	c15 := pCnt * (cPerc - 10)

	loads := map[string]map[string]string{
		"l1": {
			"oid":   laLoadInt + ".1",
			"name":  "load_1_min",
			"warn":  strconv.Itoa(w1),
			"crit":  strconv.Itoa(c1),
			"wReal": fmt.Sprintf("%.2f", float64(w1)/100),
			"cReal": fmt.Sprintf("%.2f", float64(c1)/100),
		},
		"l5": {
			"oid":   laLoadInt + ".2",
			"name":  "load_5_min",
			"warn":  strconv.Itoa(w5),
			"crit":  strconv.Itoa(c5),
			"wReal": fmt.Sprintf("%.2f", float64(w5)/100),
			"cReal": fmt.Sprintf("%.2f", float64(c5)/100),
		},
		"l15": {
			"oid":   laLoadInt + ".3",
			"name":  "load_15_min",
			"warn":  strconv.Itoa(w15),
			"crit":  strconv.Itoa(c15),
			"wReal": fmt.Sprintf("%.2f", float64(w15)/100),
			"cReal": fmt.Sprintf("%.2f", float64(c15)/100),
		},
	}

	// Do SNMP query
	res, err = l.Sess.Get([]string{loads["l1"]["oid"], loads["l5"]["oid"], loads["l15"]["oid"]})
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	l.Check.AddMsg(0, fmt.Sprintf("%d CPUs", pCnt), "")

	for _, p := range [3]string{"l1", "l5", "l15"} {
		v := res[loads[p]["oid"]].Integer
		level, err := l.Check.AlarmLevel(v, loads[p]["warn"], loads[p]["crit"])
		if err != nil {
			return fmt.Errorf("alarm level error: %v", err)
		}

		vReal := fmt.Sprintf("%.2f", float64(v)/100)
		l.Check.AddPerfData(loads[p]["name"], vReal, "", loads[p]["wReal"], loads[p]["cReal"], "0", "")
		l.Check.AddMsg(level, fmt.Sprintf("%s %s", p, vReal), "")
	}

	return nil
}

// Get Juniper load data using jnxOperatingTable
func (l *Load) jnxLoad() error {
	// Find routing engines
	res, err := l.Sess.Walk(jnxOperatingDescr, true, true)
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	re := make(map[string]string)
	for i, d := range res {
		if strings.Contains(strings.ToUpper(d.OctetString), strings.ToUpper("Routing Engine")) {
			re[i] = d.OctetString
		}
	}

	loads := make(map[string]map[string]uint64)
	for i, n := range re {
		// Do SNMP query
		o := []string{jnxOperatingCPU + "." + i, jnxOperating1MinLoadAvg + "." + i, jnxOperating5MinLoadAvg + "." + i}
		res, err := l.Sess.Get(o)
		if err != nil {
			return fmt.Errorf("snmp error: %v", err)
		}
		// DEBUG
		if l.Debug {
			fmt.Printf("%# v\n", pretty.Formatter(res))
		}

		d := map[string]uint64{
			"util":  res[jnxOperatingCPU+"."+i].Gauge32,
			"load1": res[jnxOperating1MinLoadAvg+"."+i].Gauge32,
			"load5": res[jnxOperating5MinLoadAvg+"."+i].Gauge32,
		}

		loads[n] = d
	}

	cn := make([]string, len(loads))
	i := 0
	for k := range loads {
		cn[i] = k
		i++
	}
	sort.Strings(cn)

	for _, n := range cn {
		l.Check.AddMsg(0, n, "")

		if v, ok := loads[n]["util"]; ok {
			level, err := l.Check.AlarmLevel(int64(v), l.Warn, l.Crit)
			if err != nil {
				return fmt.Errorf("alarm level error: %v", err)
			}
			l.Check.AddPerfData("'"+n+" util'", fmt.Sprintf("%d", v), "%", l.Warn, l.Crit, "0", "")
			l.Check.AddMsg(level, fmt.Sprintf("util %d%%", v), "")
		} else {
			l.Check.AddMsg(3, "util Na", "")
		}

		for _, t := range []string{"1", "5"} {
			if v, ok := loads[n]["load"+t]; ok {
				l.Check.AddPerfData("'"+n+" load"+t+"'", fmt.Sprintf("%d", v), "%", "", "", "0", "")
				l.Check.AddMsg(0, fmt.Sprintf("load%s %d%%", t, v), "")
			} else {
				l.Check.AddMsg(3, "load"+t+" Na", "")
			}
		}
	}

	return nil
}

// Get Cisco load data using ciscoProcessMIB
func (l *Load) ciscoLoad() error {
	// Find CPU entity id-s
	res, err := l.Sess.Walk(cpmCPUTotalPhysicalIndex, true, true)
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	names := make(map[string]string)
	cpuIDs := make(map[string]int64)
	for i, d := range res {
		if d.Integer == 0 {
			names[i] = "CPU0"
			continue
		}
		cpuIDs[i] = d.Integer
	}

	// Find entity names
	eo := make([]string, len(cpuIDs))
	i := 0
	for _, v := range cpuIDs {
		eo[i] = fmt.Sprintf("%s.%d", entPhysicalName, v)
		i++
	}

	res, err = l.Sess.Get(eo)
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	for idx, eidx := range cpuIDs {
		oid := fmt.Sprintf("%s.%d", entPhysicalName, eidx)
		if res[oid].OctetString != "" {
			names[idx] = res[oid].OctetString
		}
	}

	// Get CPU load data
	lo := make([]string, 2*len(names))
	i = 0
	for idx := range names {
		lo[i] = cpmCPUTotal1minRev + "." + idx
		i++
		lo[i] = cpmCPUTotal5minRev + "." + idx
		i++
	}

	res, err = l.Sess.Get(lo)
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	loads := make(map[string]map[string]uint64)
	for idx, n := range names {
		l1mo := cpmCPUTotal1minRev + "." + idx
		l5mo := cpmCPUTotal5minRev + "." + idx

		d := make(map[string]uint64)
		if v, ok := res[l1mo]; ok {
			d["l1m"] = v.Gauge32
		}
		if v, ok := res[l5mo]; ok {
			d["l5m"] = v.Gauge32
		}
		loads[n] = d
	}

	wInt, err := strconv.Atoi(l.Warn)
	if err != nil {
		return fmt.Errorf("warning level must be integer: %v", err)
	}

	cInt, err := strconv.Atoi(l.Crit)
	if err != nil {
		return fmt.Errorf("critical level must be integer: %v", err)
	}

	// Calculate alarm levels for 5 min values
	w5m := strconv.Itoa(wInt - 5)
	c5m := strconv.Itoa(cInt - 5)

	cn := make([]string, len(loads))
	i = 0
	for k := range loads {
		cn[i] = k
		i++
	}
	sort.Strings(cn)

	for _, n := range cn {
		l.Check.AddMsg(0, n, "")

		if v, ok := loads[n]["l1m"]; ok {
			level, err := l.Check.AlarmLevel(int64(v), l.Warn, l.Crit)
			if err != nil {
				return fmt.Errorf("alarm level error: %v", err)
			}
			l.Check.AddPerfData("'"+n+" 1min'", fmt.Sprintf("%d", v), "%", l.Warn, l.Crit, "0", "")
			l.Check.AddMsg(level, fmt.Sprintf("1m %d%%", v), "")
		} else {
			l.Check.AddMsg(3, "1m Na", "")
		}

		if v, ok := loads[n]["l5m"]; ok {
			level, err := l.Check.AlarmLevel(int64(v), w5m, c5m)
			if err != nil {
				return fmt.Errorf("alarm level error: %v", err)
			}
			l.Check.AddPerfData("'"+n+" 5min'", fmt.Sprintf("%d", v), "%", w5m, c5m, "0", "")
			l.Check.AddMsg(level, fmt.Sprintf("5m %d%%", v), "")
		} else {
			l.Check.AddMsg(3, "5m Na", "")
		}

		l.Check.AddPerfData("dummy", "0", "", "", "", "", "")
	}

	return nil
}

// Get load data using rcDeviceStsCpuUsagePercent oid
func (l *Load) ruggedSwLoad() error {
	// Do SNMP query
	res, err := l.Sess.Get([]string{rcDeviceStsCpuUsagePercent})
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	u := int64(res[rcDeviceStsCpuUsagePercent].Integer)

	level, err := l.Check.AlarmLevel(u, l.Warn, l.Crit)
	if err != nil {
		return fmt.Errorf("alarm level error: %v", err)
	}

	l.Check.AddPerfData("cpu_usage", fmt.Sprintf("%d", u), "%", l.Warn, l.Crit, "0", "100")
	l.Check.AddPerfData("dummy1", "0", "", "", "", "", "")
	l.Check.AddPerfData("dummy2", "0", "", "", "", "", "")
	l.Check.AddMsg(level, fmt.Sprintf("usage %d%%", u), "")

	return nil
}

// Get Moxa load data using cpuLoading5s cpuLoading30s cpuLoading300s oids
func (l *Load) moxaSwLoad() error {
	// Get sysobjectid
	res, err := l.Sess.Get([]string{sysObjectID})
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	soi := res[sysObjectID].ObjectIdentifier

	ol5 := soi + ".1.53.0"
	ol30 := soi + ".1.54.0"
	ol300 := soi + ".1.55.0"

	res, err = l.Sess.Get([]string{ol5, ol30, ol300})
	if err != nil {
		return fmt.Errorf("snmp error: %v", err)
	}
	// DEBUG
	if l.Debug {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}

	l5 := int64(res[ol5].Integer)
	l30 := int64(res[ol30].Integer)
	l300 := int64(res[ol300].Integer)

	wInt, err := strconv.Atoi(l.Warn)
	if err != nil {
		return fmt.Errorf("warning level must be integer: %v", err)
	}

	cInt, err := strconv.Atoi(l.Crit)
	if err != nil {
		return fmt.Errorf("critical level must be integer: %v", err)
	}

	// Calculate alarm levels for 30s and 300s values
	w30s := strconv.Itoa(wInt - 5)
	c30s := strconv.Itoa(cInt - 5)
	w300s := strconv.Itoa(wInt - 10)
	c300s := strconv.Itoa(cInt - 10)

	level, err := l.Check.AlarmLevel(l5, l.Warn, l.Crit)
	if err != nil {
		return fmt.Errorf("alarm level error: %v", err)
	}
	l.Check.AddPerfData("usage_5s", fmt.Sprintf("%d", l5), "%", l.Warn, l.Crit, "0", "100")
	l.Check.AddMsg(level, fmt.Sprintf("usage 5s %d%%", l5), "")

	level, err = l.Check.AlarmLevel(l30, w30s, c30s)
	if err != nil {
		return fmt.Errorf("alarm level error: %v", err)
	}
	l.Check.AddPerfData("usage_30s", fmt.Sprintf("%d", l30), "%", w30s, c30s, "0", "100")
	l.Check.AddMsg(level, fmt.Sprintf("30s %d%%", l30), "")

	level, err = l.Check.AlarmLevel(l300, w300s, c300s)
	if err != nil {
		return fmt.Errorf("alarm level error: %v", err)
	}
	l.Check.AddPerfData("usage_300s", fmt.Sprintf("%d", l300), "%", w300s, c300s, "0", "100")
	l.Check.AddMsg(level, fmt.Sprintf("300s %d%%", l300), "")

	return nil
}

// Returns load data as cpu cnt and load map
func calcCPUData(data snmphelper.SnmpOut) (map[string]int64, error) {
	var loads []int64

	for _, d := range data {
		loads = append(loads, d.Integer)
	}

	cnt := int64(len(loads))
	if cnt == 0 {
		return nil, fmt.Errorf("CPU count 0 or unknown")
	}

	var loadSum int64 = 0
	for _, v := range loads {
		loadSum += v
	}

	var loadAvg float64 = float64(loadSum) / float64(cnt)
	var load int64 = int64(math.Round(loadAvg))

	out := map[string]int64{"cpuCnt": cnt, "load": load}

	return out, nil
}
