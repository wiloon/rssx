module rssx

go 1.14

require (
	github.com/garyburd/redigo v1.6.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/satori/go.uuid v1.2.0
	github.com/wiloon/pingd-config v0.0.0-20190908085236-59c3745180bc
	github.com/wiloon/pingd-data v0.0.0-20190824105510-017ed144fa34
	github.com/wiloon/pingd-log v1.0.2
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/wiloon/pingd-log v1.0.2 => /home/wiloon/projects/pingd-log
replace github.com/wiloon/pingd-config v0.0.0-20190908085236-59c3745180bc => /home/wiloon/projects/pingd-config
replace github.com/wiloon/pingd-data v0.0.0-20190824105510-017ed144fa34 => /home/wiloon/projects/pingd-data

