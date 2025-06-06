package templates

import "github.com/donseba/go-form/types"

var TailwindV3 = map[types.FieldType]map[types.InputFieldType]string{
	types.FieldTypeBase: {
		types.InputFieldTypeNone: `<input type="{{.Type}}" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{ form_print .Loc .Field.Placeholder}}" {{if .Field.Required}}required{{end}} class="border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>`,
	},
	types.FieldTypeInput: {
		types.InputFieldTypeNone:          `{{ baseInput "Type" "text" "Field" .Field}}`,
		types.InputFieldTypeText:          `{{ baseInput "Type" "text" "Field" .Field}}`,
		types.InputFieldTypePassword:      `{{ baseInput "Type" "password" "Field" .Field}}`,
		types.InputFieldTypeEmail:         `{{ baseInput "Type" "email" "Field" .Field}}`,
		types.InputFieldTypeTel:           `{{ baseInput "Type" "tel" "Field" .Field}}`,
		types.InputFieldTypeNumber:        `<input type="number" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{if .Field.Min}}min="{{.Field.Min}}"{{end}} {{if .Field.Max}}max="{{.Field.Max}}"{{end}} {{if .Field.Step}}step="{{.Field.Step}}"{{end}} {{if .Field.Required}}required{{end}} class="border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>`,
		types.InputFieldTypeDate:          `{{ baseInput "Type" "date" "Field" .Field}}`,
		types.InputFieldTypeDateTimeLocal: `{{ baseInput "Type" "datetime-local" "Field" .Field}}`,
		types.InputFieldTypeTime:          `{{ baseInput "Type" "time" "Field" .Field}}`,
		types.InputFieldTypeWeek:          `{{ baseInput "Type" "week" "Field" .Field}}`,
		types.InputFieldTypeMonth:         `{{ baseInput "Type" "month" "Field" .Field}}`,
		types.InputFieldTypeSearch:        `{{ baseInput "Type" "search" "Field" .Field}}`,
		types.InputFieldTypeUrl:           `{{ baseInput "Type" "url" "Field" .Field}}`,
		types.InputFieldTypeColor: `<div class="flex items-center gap-2">
  <input type="color" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" class="h-8 w-8 rounded border border-gray-200 p-0" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  <span class="text-sm text-gray-500" aria-hidden="true">{{.Field.Value}}</span>
</div>`,
		types.InputFieldTypeRange: `<div class="flex items-center gap-2">
  <input type="range" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{if .Field.Min}}min="{{.Field.Min}}"{{end}} {{if .Field.Max}}max="{{.Field.Max}}"{{end}} {{if .Field.Step}}step="{{.Field.Step}}"{{end}} class="w-full h-2 rounded-lg appearance-none cursor-pointer bg-gray-200 border border-gray-200" oninput="document.getElementById('{{.Field.Id}}_value').textContent = this.value" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  <span id="{{.Field.Id}}_value" class="text-sm text-gray-500 min-w-[3rem] text-right" aria-hidden="true">{{.Field.Value}}</span>
</div>`,
		types.InputFieldTypeImage:  `<input type="image" id="{{.Field.Id}}" name="{{.Field.Name}}" src="{{.Field.Value}}" alt="{{.Field.Label}}" class="border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>`,
		types.InputFieldTypeSubmit: `<button type="submit" class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed" {{ if eq .Field.Disabled true }}disabled{{end}} aria-labelledby="{{.Field.Id}}_label">{{ form_print .Loc .Field.Label }}</button>`,
		types.InputFieldTypeHidden: `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}">`,
	},
	types.FieldTypeCheckbox: {
		types.InputFieldTypeNone: `<div class="inline-block">
  <input type="checkbox" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} class="h-4 w-4 rounded border border-gray-200 text-indigo-600 focus:ring-indigo-600" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
</div>`,
	},
	types.FieldTypeRadios: {
		types.InputFieldTypeNone: `<div role="radiogroup" aria-labelledby="{{.Field.Id}}_label">
  {{ range $k, $option := .Field.Values }}
  <div class="inline-block mr-4">
    <input type="radio" id="{{$.Field.Id}}_{{$k}}" name="{{$.Field.Name}}" value="{{$option.Value}}" {{ if eq $.Field.Value $option.Value }}checked{{end}} {{ if eq $.Field.Required true }}required{{end}} class="h-4 w-4 border border-gray-200 text-indigo-600 focus:ring-indigo-600" aria-labelledby="{{$.Field.Id}}_{{$k}}_label" {{if $.Field.Description}}aria-describedby="{{$.Field.Id}}_description"{{end}}>
    <label for="{{$.Field.Id}}_{{$k}}" id="{{$.Field.Id}}_{{$k}}_label" class="ml-2 text-sm text-gray-900">{{$option.Name}}</label>
  </div>
  {{ end }}
</div>`,
		types.InputFieldTypeRadioStruct: `<div class="inline-block">
  <input type="radio" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Id}}" {{ if eq .Field.Value .Field.Id }}checked{{end}} {{ if eq .Field.Required true }}required{{end}} class="h-4 w-4 border border-gray-200 text-indigo-600 focus:ring-indigo-600" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
</div>`,
	},
	types.FieldTypeDropdown: {
		types.InputFieldTypeNone: `<select id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} class="border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeDropdownMapped: {
		types.InputFieldTypeNone: `<select id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} class="border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeTextArea: {
		types.InputFieldTypeNone: `<div class="relative">
  <textarea id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} class="border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>{{.Field.Value}}</textarea>
</div>`,
	},
	types.FieldTypeGroup: {
		types.InputFieldTypeNone: `<div class="mb-2 rounded-lg border border-gray-200 bg-white">
  <div class="border-b border-gray-200 bg-gray-50 px-4 py-2">
    <h6 class="m-0 text-sm font-medium text-gray-900" id="{{.Field.Id}}_legend">{{.Field.Legend}}</h6>
  </div>
  <div class="p-4" role="group" aria-labelledby="{{.Field.Id}}_legend">
    {{ fields }}
  </div>
</div>`,
	},
	types.FieldTypeError: {
		types.InputFieldTypeNone: `{{range errors}}<div class="mt-1 text-sm text-red-600" role="alert">{{.}}</div>{{end}}`,
	},
	types.FieldTypeLabel: {
		types.InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} class="block text-sm font-medium leading-6 text-gray-900" {{with .Field.Id}}id="{{.}}_label"{{end}}>{{ form_print .Loc .Field.Label}}{{ if eq .Field.Required true }}<span class="text-red-500" aria-hidden="true">*</span><span class="sr-only">(required)</span>{{end}}</label>`,
	},
	types.FieldTypeWrapper: {
		types.InputFieldTypeNone: `<div class="mb-2">
  {{ label }}
  {{ field }}
  {{ range errors }}
  <div class="mt-1 text-sm text-red-600" role="alert">{{ . }}</div>
  {{ end }}
  {{ if .Field.Description }}
  <div class="mt-1 text-sm text-gray-500" id="{{.Field.Id}}_description">{{.Field.Description}}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeForm: {
		types.InputFieldTypeNone: `<form action="{{.Field.Target}}" method="{{.Field.Method}}" class="mx-auto max-w-md rounded-lg border border-gray-200 bg-white p-4 shadow-sm" {{ if .Field.Attributes }}{{ form_attributes .Field.Attributes }}{{end}}>
  {{ fields }}
  <div class="mt-4 flex justify-end">
    <button type="submit" class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed">{{ form_print .Loc .Field.Label }}</button>
  </div>
</form>`,
	},
}
