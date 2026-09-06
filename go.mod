module webtyp.com/app-demo

go 1.25.2

require (
	webtyp.com/components v0.6.16
	webtyp.com/css v0.4.21
	webtyp.com/dom v0.13.10
	webtyp.com/fmt v1.0.0
	webtyp.com/html v0.0.21
	webtyp.com/input v0.0.6
	webtyp.com/layout v0.2.16
	webtyp.com/model v0.1.8
	webtyp.com/orm v0.12.1
	webtyp.com/storage v0.0.7
	webtyp.com/svg v0.3.5
	webtyp.com/time v0.5.5
	webtyp.com/unixid v0.2.28
	webtyp.com/view v0.5.2
)

require (
	webtyp.com/color v0.1.2 // indirect
	webtyp.com/date v0.0.6 // indirect
	webtyp.com/font v0.0.5 // indirect
	webtyp.com/form v0.4.7 // indirect
	webtyp.com/icons v0.0.3 // indirect
	webtyp.com/json v0.5.25 // indirect
	webtyp.com/router v0.1.31 // indirect
	webtyp.com/widget v0.6.24 // indirect
)

// Local replaces for unreleased work: the daemon serves these live, so no
// publish is needed to verify in the running demo. Drop each line once its
// repo is published past the change.

replace webtyp.com/components => ../components

replace webtyp.com/icons => ../icons
