package templates

import "github.com/donseba/go-form/types"

// TailwindV4 targets Tailwind CSS v4 projects.
// It uses modern utility patterns (e.g. focus-visible) and stays generic (no custom CSS tokens).
var TailwindV4 = map[types.FieldType]map[types.InputFieldType]string{
	types.FieldTypeBase: {
		types.InputFieldTypeNone: `<input type="{{.Type}}" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{ form_print .Loc .Field.Placeholder}}" {{if .Field.Required}}required{{end}} {{if .Field.Min}}min="{{.Field.Min}}"{{end}} {{if .Field.Max}}max="{{.Field.Max}}"{{end}} {{if .Field.Step}}step="{{.Field.Step}}"{{end}} class="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm placeholder:text-gray-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder:text-gray-400 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900 {{ .Field.Class }}" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}} {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>`,
	},
	types.FieldTypeInputGroup: {
		types.InputFieldTypeNone: `<div class="flex w-full rounded-md shadow-sm {{ .Field.Class }}">
  {{if .GroupBefore}}<span class="inline-flex items-center rounded-l-md border border-r-0 border-gray-300 bg-gray-50 px-3 text-sm text-gray-500 dark:border-gray-700 dark:bg-gray-700 dark:text-gray-200">{{.GroupBefore}}</span>{{end}}
  {{.Input}}
  {{if .GroupAfter}}<span class="inline-flex items-center rounded-r-md border border-l-0 border-gray-300 bg-gray-50 px-3 text-sm text-gray-500 dark:border-gray-700 dark:bg-gray-700 dark:text-gray-200">{{.GroupAfter}}</span>{{end}}
</div>`,
	},
	types.FieldTypeInput: {
		types.InputFieldTypeNone:          `{{ baseInput "Type" "text" "Field" .Field}}`,
		types.InputFieldTypeText:          `{{ baseInput "Type" "text" "Field" .Field}}`,
		types.InputFieldTypePassword:      `{{ baseInput "Type" "password" "Field" .Field}}`,
		types.InputFieldTypeEmail:         `{{ baseInput "Type" "email" "Field" .Field}}`,
		types.InputFieldTypeTel:           `{{ baseInput "Type" "tel" "Field" .Field}}`,
		types.InputFieldTypeNumber:        `<input type="number" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{if .Field.Min}}min="{{.Field.Min}}"{{end}} {{if .Field.Max}}max="{{.Field.Max}}"{{end}} {{if .Field.Step}}step="{{.Field.Step}}"{{end}} {{if .Field.Required}}required{{end}} class="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm placeholder:text-gray-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder:text-gray-400 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900 {{ .Field.Class }}" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}} {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>`,
		types.InputFieldTypeDate:          `{{ baseInput "Type" "date" "Field" .Field}}`,
		types.InputFieldTypeDateTimeLocal: `{{ baseInput "Type" "datetime-local" "Field" .Field}}`,
		types.InputFieldTypeTime:          `{{ baseInput "Type" "time" "Field" .Field}}`,
		types.InputFieldTypeWeek:          `{{ baseInput "Type" "week" "Field" .Field}}`,
		types.InputFieldTypeMonth:         `{{ baseInput "Type" "month" "Field" .Field}}`,
		types.InputFieldTypeSearch:        `{{ baseInput "Type" "search" "Field" .Field}}`,
		types.InputFieldTypeUrl:           `{{ baseInput "Type" "url" "Field" .Field}}`,
		types.InputFieldTypeColor: `<div class="flex items-center gap-2">
  <input type="color" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" class="h-8 w-8 rounded border border-gray-300 bg-white p-0 dark:border-gray-700 dark:bg-gray-800 {{ .Field.Class }}" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}} {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>
  <span class="text-sm text-gray-600 dark:text-gray-300" aria-hidden="true">{{.Field.Value}}</span>
</div>`,
		types.InputFieldTypeRange: `<div class="flex items-center gap-2">
  <input type="range" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{if .Field.Min}}min="{{.Field.Min}}"{{end}} {{if .Field.Max}}max="{{.Field.Max}}"{{end}} {{if .Field.Step}}step="{{.Field.Step}}"{{end}} class="h-2 w-full cursor-pointer appearance-none rounded-lg bg-gray-200 dark:bg-gray-700 {{ .Field.Class }}" oninput="document.getElementById('{{.Field.Id}}_value').textContent = this.value" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}} {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>
  <span id="{{.Field.Id}}_value" class="min-w-[3rem] text-right text-sm text-gray-600 dark:text-gray-300" aria-hidden="true">{{.Field.Value}}</span>
</div>`,
		types.InputFieldTypeImage:  `<input type="image" id="{{.Field.Id}}" name="{{.Field.Name}}" src="{{.Field.Value}}" alt="{{.Field.Label}}" class="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900 {{ .Field.Class }}" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}} {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>`,
		types.InputFieldTypeSubmit: `<button type="submit" class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white disabled:cursor-not-allowed disabled:opacity-50 dark:bg-indigo-500 dark:hover:bg-indigo-400 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900 {{ .Field.Class }}" {{ if eq .Field.Disabled true }}disabled{{end}} aria-labelledby="{{.Field.Id}}_label">{{ form_print .Loc .Field.Label }}</button>`,
		types.InputFieldTypeHidden: `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}" {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>`,
	},
	types.FieldTypeCheckbox: {
		types.InputFieldTypeNone: `<div class="inline-block">
  <input type="checkbox" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} class="h-4 w-4 rounded border border-gray-300 text-indigo-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:border-gray-700 dark:text-indigo-400 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900 {{ .Field.Class }}" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}} {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>
</div>`,
	},
	types.FieldTypeRadios: {
		types.InputFieldTypeNone: `<div role="radiogroup" aria-labelledby="{{.Field.Id}}_label" class="{{ .Field.Class }}">
  {{ range $k, $option := .Field.Values }}
  <div class="inline-block mr-4">
    <input type="radio" id="{{$.Field.Id}}_{{$k}}" name="{{$.Field.Name}}" value="{{$option.Value}}" {{ if eq $.Field.Value $option.Value }}checked{{end}} {{ if eq $.Field.Required true }}required{{end}} class="h-4 w-4 border border-gray-300 text-indigo-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:border-gray-700 dark:text-indigo-400 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900" aria-labelledby="{{$.Field.Id}}_{{$k}}_label" {{if $.Field.Description}}aria-describedby="{{$.Field.Id}}_description"{{end}}>
    <label for="{{$.Field.Id}}_{{$k}}" id="{{$.Field.Id}}_{{$k}}_label" class="ml-2 text-sm text-gray-900 dark:text-gray-100">{{$option.Name}}</label>
  </div>
  {{ end }}
</div>`,
		types.InputFieldTypeRadioStruct: `<div class="inline-block">
  <input type="radio" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Id}}" {{ if eq .Field.Value .Field.Id }}checked{{end}} {{ if eq .Field.Required true }}required{{end}} class="h-4 w-4 border border-gray-300 text-indigo-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:border-gray-700 dark:text-indigo-400 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900 {{ .Field.Class }}" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
</div>`,
	},
	types.FieldTypeDropdown: {
		types.InputFieldTypeNone: `<select id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900 {{ .Field.Class }}" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}} {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeDropdownMapped: {
		types.InputFieldTypeNone: `<select id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900 {{ .Field.Class }}" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}} {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeTextArea: {
		types.InputFieldTypeNone: `<div class="relative">
  <textarea id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}" placeholder="{{ form_print .Loc .Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} class="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm placeholder:text-gray-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder:text-gray-400 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900 {{ .Field.Class }}" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}} {{if .Field.Data}}{{ form_data_attributes .Field.Data }}{{end}}>{{.Field.Value}}</textarea>
</div>`,
	},
	types.FieldTypeGroup: {
		types.InputFieldTypeNone: `<div class="mb-2 rounded-md border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-gray-800 {{ .Field.Class}}">
  <div class="border-b border-gray-200 px-4 py-2 dark:border-gray-700">
    <h6 class="m-0 text-sm font-medium text-gray-900 dark:text-gray-100" id="{{.Field.Id}}_legend">{{ form_print .Loc .Field.Legend}}</h6>
  </div>
  <div class="p-4" role="group" aria-labelledby="{{.Field.Id}}_legend">
    {{ fields }}
  </div>
</div>`,
	},
	types.FieldTypeError: {
		types.InputFieldTypeNone: `{{range errors}}<div class="mt-1 text-sm text-red-600 dark:text-red-400" role="alert">{{.}}</div>{{end}}`,
	},
	types.FieldTypeLabel: {
		types.InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} class="block text-sm font-medium leading-6 text-gray-900 dark:text-gray-100" {{with .Field.Id}}id="{{.}}_label"{{end}}>{{ form_print .Loc .Field.Label}}{{ if eq .Field.Required true }}<span class="text-red-600 dark:text-red-400" aria-hidden="true">*</span><span class="sr-only">(required)</span>{{end}}</label>`,
	},
	types.FieldTypeWrapper: {
		types.InputFieldTypeNone: `<div class="mb-2">
  {{ label }}
  {{ field }}
  {{ range errors }}
  <div class="mt-1 text-sm text-red-600 dark:text-red-400" role="alert">{{ . }}</div>
  {{ end }}
  {{ if .Field.Description }}
  <div class="mt-1 text-sm text-gray-500 dark:text-gray-300" id="{{.Field.Id}}_description">{{.Field.Description}}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeForm: {
		types.InputFieldTypeNone: `<form action="{{.Field.Target}}" method="{{.Field.Method}}" class="mx-auto max-w-md rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-800 {{ .Field.Class }}" {{ if .Field.Attributes }}{{ form_attributes .Field.Attributes }}{{end}}>
  {{ fields }}
  <div class="mt-4 flex justify-end gap-2">
    {{ if .Field.CancelTarget }}
      <a href="{{ .Field.CancelTarget }}" class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:bg-gray-800 dark:text-gray-100 dark:ring-gray-700 dark:hover:bg-gray-700 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900">{{ if .Field.CancelText }}{{ form_print .Loc .Field.CancelText }}{{ else }}{{ form_print .Loc "Cancel" }}{{ end }}</a>
    {{ end }}
    <button type="submit" class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 focus-visible:ring-offset-white disabled:cursor-not-allowed disabled:opacity-50 dark:bg-indigo-500 dark:hover:bg-indigo-400 dark:focus-visible:ring-indigo-500 dark:focus-visible:ring-offset-gray-900">{{ form_print .Loc .Field.Label }}</button>
  </div>
</form>`,
	},
}
