module github.com/aretaja/check-gosnmp-cpu

go 1.15

// For local development
replace (
	github.com/aretaja/check-gosnmp-cpu => ./
	github.com/aretaja/icingahelper => ../icingahelper
	github.com/aretaja/snmphelper => ../snmphelper
)

require (
	github.com/aretaja/icingahelper v1.1.1
	github.com/aretaja/snmphelper v1.1.3
	github.com/gosnmp/gosnmp v1.36.1 // indirect
	github.com/kr/pretty v0.3.1
	github.com/rogpeppe/go-internal v1.11.0 // indirect
)
