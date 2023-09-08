module github.com/aretaja/check-gosnmp-cpu

go 1.15

// For local development
replace (
	github.com/aretaja/check-gosnmp-cpu => ./
	github.com/aretaja/icingahelper => ../icingahelper
	github.com/aretaja/snmphelper => ../snmphelper
)

require (
	github.com/aretaja/icingahelper v1.0.0
	github.com/aretaja/snmphelper v1.0.0
	github.com/kr/pretty v0.2.1
)
