package templates

import "github.com/donseba/go-form/types"

var TailwindV3 = map[types.FieldType]map[types.InputFieldType]string{
	types.FieldTypeInput: {
		types.InputFieldTypeText:     `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="text" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-indigo-500 dark:focus:ring-indigo-500">`,
		types.InputFieldTypePassword: `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="password" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-indigo-500 dark:focus:ring-indigo-500">`,
		types.InputFieldTypeEmail:    `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="email" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-indigo-500 dark:focus:ring-indigo-500">`,
		types.InputFieldTypeTel:      `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="tel" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-indigo-500 dark:focus:ring-indigo-500">`,
		types.InputFieldTypeNumber:   `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="number" step="{{.Field.Step}}" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-indigo-500 dark:focus:ring-indigo-500">`,
		types.InputFieldTypeDate:     `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="date" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-indigo-500 dark:focus:ring-indigo-500">`,
		types.InputFieldTypeNone:     `<input id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-indigo-500 dark:focus:ring-indigo-500">`,
		types.InputFieldTypeHidden:   `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
	},
	types.FieldTypeCheckbox: {
		types.InputFieldTypeNone: `<div class="inline-flex items-center">
  <input id="{{.Field.Id}}" name="{{.Field.Name}}" type="checkbox" {{ if eq .Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:ring-offset-gray-800 dark:focus:ring-indigo-500">
</div>`,
	},
	types.FieldTypeRadios: {
		types.InputFieldTypeNone: `{{ range $k, $option := .Field.Values }}
<div class="inline-flex items-center mr-4">
  <input id="{{$.Field.Id}}_{{$k}}" name="{{$.Field.Name}}" type="radio" value="{{$option.Value}}" {{ if eq $.Field.Value $option.Value }}checked{{end}} {{ if eq $.Field.Required true }}required{{end}} class="h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:ring-offset-gray-800 dark:focus:ring-indigo-500">
  <label for="{{$.Field.Id}}_{{$k}}" class="ml-2 text-sm text-gray-900 dark:text-gray-300">{{$option.Name}}</label>
</div>
{{ end }}`,
		types.InputFieldTypeRadioStruct: `<div class="inline-flex items-center">
  <input id="{{.Field.Id}}" name="{{.Field.Name}}" type="radio" value="{{.Field.Id}}" {{ if eq .Field.Value .Field.Id }}checked{{end}} {{ if eq .Field.Required true }}required{{end}} class="h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:ring-offset-gray-800 dark:focus:ring-indigo-500">
</div>`,
	},
	types.FieldTypeDropdown: {
		types.InputFieldTypeNone: `<select {{with .Field.Id}}id="{{.}}"{{end}} name="{{.Field.Name}}" class="block w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-indigo-500 dark:focus:ring-indigo-500">
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeTextArea: {
		types.InputFieldTypeNone: `<textarea id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-indigo-500 dark:focus:ring-indigo-500">{{.Field.Value}}</textarea>`,
	},
	types.FieldTypeGroup: {
		types.InputFieldTypeNone: `<div class="rounded-lg border border-gray-200 bg-white p-2 shadow-sm dark:border-gray-700 dark:bg-gray-800 dark:shadow-md">
  <h6 class="mb-2 text-sm font-medium text-gray-900 dark:text-white">{{.Field.Legend}}</h6>
  <div class="space-y-2">
    {{ fields }}
  </div>
</div>`,
	},
	types.FieldTypeError: {
		types.InputFieldTypeNone: `{{range errors}}<div class="mt-1 text-xs text-red-600 dark:text-red-400">{{.}}</div>{{end}}`,
	},
	types.FieldTypeLabel: {
		types.InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{.Field.Label}}{{ if eq .Field.Required true }}<span class="text-red-500 dark:text-red-400">*</span>{{end}}</label>`,
	},
	types.FieldTypeWrapper: {
		types.InputFieldTypeNone: `<div class="mb-2">
  {{ label }}
  {{ field }}
  {{ range errors }}
  <div class="mt-1 text-xs text-red-600 dark:text-red-400">{{ . }}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeForm: {
		types.InputFieldTypeNone: `<form action="{{.Field.Target}}" method="{{.Field.Method}}" class="mx-auto max-w-md rounded border border-gray-200 bg-white p-3 shadow-sm dark:border-gray-700 dark:bg-gray-800 dark:shadow-md">
  {{ fields }}
  <div class="mt-3">
    <button type="submit" class="w-full rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:bg-indigo-500 dark:hover:bg-indigo-600 dark:focus:ring-offset-gray-800">Submit</button>
  </div>
</form>`,
	},
}
