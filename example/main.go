package main

import (
	"fmt"
	"github.com/donseba/go-form"
	"html/template"
	"net/http"
)

var inputTpl = `<div>
  <label {{with .Id}}for="{{.}}"{{end}} class="block text-sm font-medium text-gray-700">{{.Label}}{{ if eq .Required true }}*{{end}}</label>
  <div class="mt-1">
	{{ if eq .Type "dropdown" }}
	<select {{with .Id}}id="{{.}}"{{end}} name="{{.Name}}" class="bg-white block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm">
	  {{ $value := .Value }}
	  {{ range $k, $option := .Values }}
		<option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
	  {{ end }}
    </select>
	{{ else if eq .Type "checkbox" }}
	<input {{with .Id}}id="{{.}}"{{end}} name="{{.Name}}" type="checkbox" {{ if eq .Required true }}required{{end}} {{ if eq .Value true }}checked{{end}} class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600">
    {{ else }}
	<input {{with .Id}}id="{{.}}"{{end}} name="{{.Name}}" placeholder="{{.Placeholder}}" {{with .Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm">
    {{ end }}
	{{range errors}}
     <span class="text-sm text-red-600">{{.}}</span>
	{{end}}
</div>
</div>`

var groupTpl = `<div class="mb-4 bg-gray-50 p-2 rounded-md">
  <label class="block text-grey-darker text-sm font-bold mb-2">
	{{.Name }}
  </label>
  {{ fields }}
</div>`

func main() {
	tpl := template.Must(template.New("").Funcs(template.FuncMap{
		"errors": func() []form.FieldError { return nil },
	}).Parse(inputTpl))
	gtpl := template.Must(template.New("").Funcs(template.FuncMap{
		"fields": func() template.HTML { return "" },
	}).Parse(groupTpl))
	fb := form.NewForm(*tpl, *gtpl)

	pageTpl := template.Must(template.New("").Funcs(fb.FuncMap()).Parse(`
<html>
  <head>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss/dist/tailwind.min.css" rel="stylesheet">
  </head>
  <body class="bg-grey-lighter">
    <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
      <div class="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
        <form class="space-y-6" action="#" method="POST">
  		{{ form_render .Form .Errors }}
  		<div class="flex items-center justify-between">
            <button type="submit" class="flex w-full justify-center rounded-md border border-transparent bg-indigo-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2">Signup</button>
          </div>
        </form>
  	</div>
    </div>
  </body>
</html>
	`))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		data := struct {
			Form   ExampleForm
			Errors []form.FieldError
		}{
			Form: ExampleForm{
				Name:  "John Wick",
				Email: "john.wick@gmail.com",
				Address: &AddressBlock{
					Street1: "121 Mill Neck",
					City:    "Long Island",
					State:   "NY",
					Zip:     "11765",
				},
				InputTypes: form.InputFieldTypeEmail,
				CheckBox:   true,
				CheckBox2:  false,
			},
			Errors: []form.FieldError{
				fieldError{
					Field: "Email",
					Issue: "is already taken",
				},
				fieldError{
					Field: "Address.Street1",
					Issue: "is required",
				},
			},
		}

		err := pageTpl.Execute(w, data)
		if err != nil {
			_, _ = fmt.Fprint(w, err)
			return
		}

	})

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}

type ExampleForm struct {
	Name       string
	Email      string `required:"true"`
	Address    *AddressBlock
	InputTypes form.InputFieldType `label:"Enum Example"`
	CheckBox   bool
	CheckBox2  bool
}

type AddressBlock struct {
	Street1 string
	City    string
	State   string
	Zip     string `label:"Postal Code"`
}

type fieldError struct {
	Field string
	Issue string
}

func (fe fieldError) Error() string {
	return fmt.Sprintf("%s:%s", fe.Field, fe.Issue)
}

func (fe fieldError) FieldError() (field, err string) {
	return fe.Field, fe.Issue
}
