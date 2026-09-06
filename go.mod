module webtyp.com/app-demo

go 1.25.2

require (
	webtyp.com/components v0.6.15
	webtyp.com/css v0.4.20
	webtyp.com/dom v0.13.9
	webtyp.com/fmt v0.25.7
	webtyp.com/html v0.0.19
	webtyp.com/input v0.0.3
	webtyp.com/layout v0.2.15
	webtyp.com/model v0.1.7
	webtyp.com/orm v0.12.0
	webtyp.com/storage v0.0.6
	webtyp.com/svg v0.3.3
	webtyp.com/time v0.5.4
	webtyp.com/unixid v0.2.26
	webtyp.com/view v0.5.1
)

require (
	webtyp.com/color v0.1.1 // indirect
	webtyp.com/date v0.0.5 // indirect
	webtyp.com/font v0.0.4 // indirect
	webtyp.com/form v0.4.0 // indirect
	webtyp.com/icons v0.0.2 // indirect
	webtyp.com/json v0.5.23 // indirect
	webtyp.com/router v0.1.30 // indirect
	webtyp.com/widget v0.6.23 // indirect
)

// Local replaces for unreleased work: the daemon serves these live, so no
// publish is needed to verify in the running demo. Drop each line once its
// repo is published past the change.

replace webtyp.com/components => ../components

replace webtyp.com/icons => ../icons
