# check-gosnmp-cpu
Icinga2 plugin designed to check CPU load
## Usage
```
$check-gosnmp-cpu -h
  -A string
        [authentication protocol pass phrase]
  -H string
        <host ip>
  -V int
        [snmp version] (1|2|3) (default 2)
  -X string
        [privacy protocol pass phrase]
  -a string
        [authentication protocol] (NoAuth|MD5|SHA)5 (default "MD5")
  -c string
        [critical level]. Look at warning level explanation (default "95")
  -d    Using this parameter will print out debug info
  -l string
        [security level] (noAuthNoPriv|authNoPriv|authPriv) (default "authPriv")
  -t string
        <check type>
                host - uses hostmib
                sysstats - uses UCD-SNMP-MIB systemStats
                loadavg - uses UCD-SNMP-MIB laTable
                jnx - uses jnxOperatingTable
                cisco - uses ciscoProcessMIB
                rcsw - uses rcDeviceStsCpuUsagePercent
                moxasw - uses moxa MIB
  -u string
        [username|community] (default "public")
  -v    Using this parameter will display the version number and exit
  -w string
        [warning level]. It depends of check type.
                host - % of average cpu utilization of all cores
                systat - % of cpu utilization
                loadavg - summary % of 1 min. load on all cores fe. 100 means that load for last min is 4 when device has 4 cpu cores
                        5 and 15 minute levels will be calculated from this value by decreasing them by 5 and 10 accordingly
                jnx - % of cpu utilization
                cisco - overall cpu busy % in the last 1 minute period
                        5 minute level will be calculated from this value by decreasing value by 5
                rcsw - % of cpu utilization
                moxasw - overall cpu busy % in the last 5 sec period
                        30 sec and 5 minute levels will be calculated from this value by decreasing value by 5 and 10 accordingly (default "85")
  -x string
        [privacy protocol] (NoPriv|DES|AES|AES192|AES256|AES192C|AES256C) (default "DES")

```
