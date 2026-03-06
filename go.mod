module github.com/aretaja/check-gosnmp-cpu

go 1.26

// For local development
replace (
	github.com/aretaja/check-gosnmp-cpu => ./
	github.com/aretaja/icingahelper => ../icingahelper
	github.com/aretaja/snmphelper => ../snmphelper
)

require (
	github.com/aretaja/icingahelper v1.1.1
	github.com/aretaja/snmphelper v1.1.3
	github.com/kr/pretty v0.3.1
)

require (
	github.com/gosnmp/gosnmp v1.43.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
)
