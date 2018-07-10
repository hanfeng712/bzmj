package dbserver

import (
	"fmt"
	"net/http"
	"text/template"
)

/*
	<table>
	<th align=center>Method</th><th align=center>Calls</th>
	{{range .Method}}
		<tr>
		<td align=left font=fixed>{{.Name}}({{.Type.ArgType}}, {{.Type.ReplyType}}) error</td>
		<td align=center>{{.Type.NumCalls}}</td>
		</tr>
	{{end}}
	</table>
*/
const debugText = `<html>
	<body>
	<title>Tables</title>
	{{range .}}
	<hr>
	Table {{.Name}}
	<hr>
	{{.State}}
	<br>
	{{.Rate}}
	{{end}}
	</body>
	</html>`

var debug = template.Must(template.New("Tables debug").Parse(debugText))

type debugHTTP struct {
	*DBServer
}

type debugService struct {
	Name  string
	State string
	Rate  string
}

type serviceArray []debugService

func (server debugHTTP) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	services := make([]debugService, 0)

	for k, v := range server.tables {
		services = append(services, debugService{k, v.tableStats.String(), v.qpsRates.String()})
	}

	err := debug.Execute(w, services)
	if err != nil {
		fmt.Fprintln(w, "table debug: error executing template:", err.Error())
	}

}
