package main

import (
	"fmt"
	"html/template"
	"os"
	"runtime"
	"time"

	"github.com/digitalocean/metis/log"
)

var (
	version   string
	revision  string
	buildDate string
	goVersion = runtime.Version()
)

var versionTmpl = template.Must(template.New("version").Parse(`
{{ .name }} - A Time Series Database

Version:     {{.version}}
Revision:    {{.revision}}
Build Date:  {{.buildDate}}
Go Version:  {{.goVersion}}
Website:     https://github.com/digitalocean/metis

Copyright (c) {{.year}} DigitalOcean, Inc. All rights reserved.

This work is licensed under the terms of the Apache 2.0 license.
For a copy, see <https://www.apache.org/licenses/LICENSE-2.0.html>.
`))

func showVersion() {
	err := versionTmpl.Execute(os.Stdout, map[string]string{
		"name":      "Metis",
		"version":   version,
		"revision":  revision,
		"buildDate": buildDate,
		"goVersion": goVersion,
		"year":      fmt.Sprintf("%d", time.Now().UTC().Year()),
	})
	if err != nil {
		log.Fatal("failed to execute version template: %+v", err)
	}
	os.Exit(0)
}
