package webreporter

import (
	"strings"
	"text/template"
)

type navigation struct {
	page *template.Template
	menu []webAnchor
}

type webAnchor struct{ URL, Name string }

///////////////////////////////////////////////////////////////////////////////

func newNavigation(page *template.Template, menu []webAnchor) *navigation {
	obj := new(navigation)

	obj.page = page
	obj.menu = menu

	return obj
}

func (obj *navigation) getMainMenu(menuItem string) string {
	w := new(strings.Builder)

	var data struct {
		MainMenu []webAnchor
	}
	data.MainMenu = make([]webAnchor, 0, len(obj.menu))

	for i := range obj.menu {
		if obj.menu[i].URL == menuItem {
			continue
		}
		data.MainMenu = append(data.MainMenu, obj.menu[i])
	}

	err := obj.page.Execute(w, data)
	checkErr(err)

	return w.String()
}

// func (obj *navigation) getSubMenu(url string, menuItems map[string]string) string {
// 	w := new(strings.Builder)
// 	sample, err := template.New("navigation2").Parse(navigationSubMenuTemplate)
// 	checkErr(err)

// 	getMenuItems := func(menuItems map[string]string) (res []struct{ Id, Name string }) {
// 		for item, value := range menuItems {
// 			res = append(res, struct {
// 				Id   string
// 				Name string
// 			}{Id: item, Name: value})
// 		}
// 		sort.Slice(res, func(i, j int) bool { return strings.Compare(res[i].Name, res[j].Name) < 0 })
// 		return
// 	}

// 	data := struct {
// 		URL       string
// 		MenuItems []struct{ Id, Name string }
// 	}{
// 		URL:       url,
// 		MenuItems: getMenuItems(menuItems),
// 	}

// 	err = sample.Execute(w, data)
// 	checkErr(err)

// 	return w.String()
// }

const navigationSubMenuTemplate = `
	{{ $url := .URL }}
	<nav class="menu">	
		<ul class="nav" style="display: inline-grid;">
		{{ range $item := .MenuItems }}
			<li><a style="width: 150px; overflow-wrap: anywhere;" 
				{{ if (eq $item.Id "") }}
				href="{{$url}}">{{$item.Name}}
				{{ else }}
				href="{{$url}}/{{$item.Id}}">{{$item.Name}}
				{{ end }}
				</a></li>
		{{end}}
		</ul>
	</nav>
`
