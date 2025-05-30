package form

var DefaultTemplates = map[FieldType]map[InputFieldType]string{
	FieldTypeInput: {
		InputFieldTypeText:     `<input type="text" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
		InputFieldTypePassword: `<input type="password" name="{{.Field.Name}}" />`,
		InputFieldTypeEmail:    `<input type="email" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
		InputFieldTypeTel:      `<input type="tel" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
		InputFieldTypeNumber:   `<input type="number" name="{{.Field.Name}}" value="{{.Field.Value}}" step="{{.Field.Step}}" />`,
		InputFieldTypeDate:     `<input type="date" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
		InputFieldTypeNone:     `<input name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
	},
	FieldTypeCheckbox: {
		InputFieldTypeNone: `<input type="checkbox" name="{{.Field.Name}}" {{if .Field.Value}}checked{{end}} />`,
	},
	FieldTypeRadios: {
		InputFieldTypeNone: `<input type="radio" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
	},
	FieldTypeDropdown: {
		InputFieldTypeNone: `<select name="{{.Field.Name}}">{{range .Field.Values}}<option value="{{.Value}}">{{.Name}}</option>{{end}}</select>`,
	},
	FieldTypeTextArea: {
		InputFieldTypeNone: `<textarea name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}">{{.Field.Value}}</textarea>`,
	},
	FieldTypeGroup: {
		InputFieldTypeNone: `<fieldset><legend>{{.Field.Legend}}</legend>{{range .Fields}}{{template "field" .}}{{end}}</fieldset>`,
	},
	FieldTypeError: {
		InputFieldTypeNone: `{{range errors}}<div>{{.}}</div>{{end}}`,
	},
	FieldTypeLabel: {
		InputFieldTypeNone: `<label {{with .Id }}for="{{.}}"{{end}} class="block text-sm font-medium text-gray-700">{{.Field.Label}}{{ if eq .Field.Required true }}*{{end}}</label>`,
	},
	FieldTypeWrapper: {
		InputFieldTypeNone: `<div>
  {{ label }}
  <div class="mt-2">
    {{ field }}
  </div>
  {{ range errors }}
  <span class="text-sm text-red-600">{{ . }}</span>
  {{ end }}
</div>`,
	},
}

var DefaultBootstrapTemplates = map[FieldType]map[InputFieldType]string{
	FieldTypeInput: {
		InputFieldTypeText:     `<input type="text" class="form-control" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
		InputFieldTypePassword: `<input type="password" class="form-control" name="{{.Field.Name}}" />`,
		InputFieldTypeEmail:    `<input type="email" class="form-control" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
		InputFieldTypeTel:      `<input type="tel" class="form-control" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
		InputFieldTypeNumber:   `<input type="number" class="form-control" name="{{.Field.Name}}" value="{{.Field.Value}}" step="{{.Field.Step}}" />`,
		InputFieldTypeDate:     `<input type="date" class="form-control" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
		InputFieldTypeNone:     `<input class="form-control" name="{{.Field.Name}}" value="{{.Value}}" />`,
	},
	FieldTypeCheckbox: {
		InputFieldTypeNone: `<input type="checkbox" class="form-check-input" name="{{.Field.Name}}" {{if .Field.Value}}checked{{end}} />`,
	},
	FieldTypeRadios: {
		InputFieldTypeNone: `<input type="radio" class="form-check-input" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
	},
	FieldTypeDropdown: {
		InputFieldTypeNone: `<select class="form-select" name="{{.Name}}">{{range .Values}}<option value="{{.Value}}">{{.Name}}</option>{{end}}</select>`,
	},
	FieldTypeTextArea: {
		InputFieldTypeNone: `<textarea class="form-control" name="{{.Name}}" rows="{{.Rows}}" cols="{{.Cols}}">{{.Value}}</textarea>`,
	},
	FieldTypeGroup: {
		InputFieldTypeNone: `<fieldset class="mb-3"><legend class="form-label">{{.Legend}}</legend>{{range .Fields}}{{template "field" .}}{{end}}</fieldset>`,
	},
	FieldTypeError: {
		InputFieldTypeNone: `{{range errors}}<div class="alert alert-danger">{{.}}</div>{{end}}`,
	},
}

var DefaultTailwindTemplates = map[FieldType]map[InputFieldType]string{
	FieldTypeInput: {
		InputFieldTypeText:     `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="text" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm">`,
		InputFieldTypePassword: `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="password" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm">`,
		InputFieldTypeEmail:    `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="tel" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm">`,
		InputFieldTypeTel:      `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="email" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm">`,
		InputFieldTypeNumber:   `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="number" step="{{.Field.Step}}" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm">`,
		InputFieldTypeDate:     `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="date" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm">`,
		InputFieldTypeNone:     `<input class="form-control" name="{{.Field.Name}}" value="{{.Value}}" />`,
		InputFieldTypeHidden:   `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
	},
	FieldTypeCheckbox: {
		InputFieldTypeNone: `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="checkbox" {{ if eq .Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600">`,
	},
	FieldTypeRadios: {
		InputFieldTypeNone: `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="radio" {{ if eq .Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600">`,
	},
	FieldTypeDropdown: {
		InputFieldTypeNone: `<select {{with .Field.Id}}id="{{.}}"{{end}} name="{{.Field.Name}}" class="bg-white block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm">
	{{ $value := .Field.Value }}
	{{ range $k, $option := .Field.Values }}
	<option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
	{{ end }}
</select>`,
	},
	FieldTypeTextArea: {
		InputFieldTypeNone: `<textarea id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}">{{.Field.Value}}</textarea>`,
	},
	FieldTypeGroup: {
		InputFieldTypeNone: `<div class="mb-4 bg-gray-50 p-2 rounded-md space-y-2">
 <label class="block text-grey-darker text-sm font-bold mb-2">
	{{.Field.Label }}
 </label>
 {{ fields }}
</div>`,
	},
	FieldTypeError: {
		InputFieldTypeNone: `{{range errors}}<div class="alert alert-danger">{{.}}</div>{{end}}`,
	},
	FieldTypeLabel: {
		InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} class="block text-sm font-medium text-gray-700">{{.Field.Label}}{{ if eq .Field.Required true }}*{{end}}</label>`,
	},

	FieldTypeWrapper: {
		InputFieldTypeNone: `<div>
  {{ label }}
  <div class="mt-2">
    {{ field }}
  </div>
  {{ range errors }}
  <span class="text-sm text-red-600">{{ . }}</span>
  {{ end }}
</div>`,
	},
}
