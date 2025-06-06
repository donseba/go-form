package templates

import "github.com/donseba/go-form/types"

var Plain = map[types.FieldType]map[types.InputFieldType]string{
	types.FieldTypeBase: {
		types.InputFieldTypeNone: `<input type="{{.Type}}" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{ form_print .Loc .Field.Placeholder}}" {{if .Field.Required}}required{{end}} {{ if .Field.MaxLength }}maxlenght="{{ .Field.MaxLength }}"{{end}} style="width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>`,
	},
	types.FieldTypeInput: {
		types.InputFieldTypeNone:          `{{ baseInput "Type" "text" "Field" .Field}}`,
		types.InputFieldTypeText:          ` {{ baseInput "Type" "text" "Field" .Field}} `,
		types.InputFieldTypePassword:      `{{ baseInput "Type" "password" "Field" .Field}}`,
		types.InputFieldTypeEmail:         `{{ baseInput "Type" "email" "Field" .Field}}`,
		types.InputFieldTypeTel:           `{{ baseInput "Type" "tel" "Field" .Field}}`,
		types.InputFieldTypeNumber:        `<input type="number" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{if .Field.Min}}min="{{.Field.Min}}"{{end}} {{if .Field.Max}}max="{{.Field.Max}}"{{end}} {{if .Field.Step}}step="{{.Field.Step}}"{{end}} {{if .Field.Required}}required{{end}} style="width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>`,
		types.InputFieldTypeDate:          `{{ baseInput "Type" "date" "Field" .Field}}`,
		types.InputFieldTypeDateTimeLocal: `{{ baseInput "Type" "datetime-local" "Field" .Field}}`,
		types.InputFieldTypeTime:          `{{ baseInput "Type" "time" "Field" .Field}}`,
		types.InputFieldTypeWeek:          `{{ baseInput "Type" "week" "Field" .Field}}`,
		types.InputFieldTypeMonth:         `{{ baseInput "Type" "month" "Field" .Field}}`,
		types.InputFieldTypeSearch:        `{{ baseInput "Type" "search" "Field" .Field}}`,
		types.InputFieldTypeUrl:           `{{ baseInput "Type" "url" "Field" .Field}}`,
		types.InputFieldTypeColor: `<div style="display: flex; align-items: center; gap: 0.5rem;">
  <input type="color" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" style="width: 2rem; height: 2rem; padding: 0; border: 1px solid #ced4da; border-radius: 0.25rem;" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  <span style="font-size: 0.875rem; color: #6c757d;" aria-hidden="true">{{.Field.Value}}</span>
</div>`,
		types.InputFieldTypeRange: `<div style="display: flex; align-items: center; gap: 0.5rem;">
  <input type="range" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{if .Field.Min}}min="{{.Field.Min}}"{{end}} {{if .Field.Max}}max="{{.Field.Max}}"{{end}} {{if .Field.Step}}step="{{.Field.Step}}"{{end}} style="width: 100%; height: 0.5rem; border: 1px solid #ced4da; border-radius: 0.25rem;" oninput="document.getElementById('{{.Field.Id}}_value').textContent = this.value" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  <span id="{{.Field.Id}}_value" style="font-size: 0.875rem; color: #6c757d; min-width: 3rem; text-align: right;" aria-hidden="true">{{.Field.Value}}</span>
</div>`,
		types.InputFieldTypeImage:  `<input type="image" id="{{.Field.Id}}" name="{{.Field.Name}}" src="{{.Field.Value}}" alt="{{.Field.Label}}" style="width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>`,
		types.InputFieldTypeSubmit: `<button type="submit" style="display: inline-block; font-weight: 400; text-align: center; white-space: nowrap; vertical-align: middle; user-select: none; border: 1px solid transparent; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; border-radius: 0.25rem; color: #fff; background-color: #0d6efd; border-color: #0d6efd; cursor: pointer; transition: color 0.15s ease-in-out, background-color 0.15s ease-in-out, border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;" {{ if eq .Field.Disabled true }}disabled{{end}} aria-labelledby="{{.Field.Id}}_label">{{ form_print .Loc .Field.Label }}</button>`,
		types.InputFieldTypeHidden: `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}">`,
	},
	types.FieldTypeCheckbox: {
		types.InputFieldTypeNone: `<div style="display: inline-block;">
  <input type="checkbox" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} style="width: 1rem; height: 1rem; margin-top: 0.25rem; vertical-align: top; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; cursor: pointer;" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
</div>`,
	},
	types.FieldTypeRadios: {
		types.InputFieldTypeNone: `<div role="radiogroup" aria-labelledby="{{.Field.Id}}_label">
  {{ range $k, $option := .Field.Values }}
  <div style="display: inline-block; margin-right: 1rem;">
    <input type="radio" id="{{$.Field.Id}}_{{$k}}" name="{{$.Field.Name}}" value="{{$option.Value}}" {{ if eq $.Field.Value $option.Value }}checked{{end}} {{ if eq $.Field.Required true }}required{{end}} style="width: 1rem; height: 1rem; margin-top: 0.25rem; vertical-align: top; background-color: #fff; border: 1px solid #ced4da; border-radius: 50%; cursor: pointer;" aria-labelledby="{{$.Field.Id}}_{{$k}}_label" {{if $.Field.Description}}aria-describedby="{{$.Field.Id}}_description"{{end}}>
    <label for="{{$.Field.Id}}_{{$k}}" id="{{$.Field.Id}}_{{$k}}_label" style="margin-left: 0.25rem; font-size: 0.875rem; color: #212529;">{{$option.Name}}</label>
  </div>
  {{ end }}
</div>`,
		types.InputFieldTypeRadioStruct: `<div style="display: inline-block;">
  <input type="radio" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Id}}" {{ if eq .Field.Value .Field.Id }}checked{{end}} {{ if eq .Field.Required true }}required{{end}} style="width: 1rem; height: 1rem; margin-top: 0.25rem; vertical-align: top; background-color: #fff; border: 1px solid #ced4da; border-radius: 50%; cursor: pointer;" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
</div>`,
	},
	types.FieldTypeDropdown: {
		types.InputFieldTypeNone: `<select id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeDropdownMapped: {
		types.InputFieldTypeNone: `<select id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeTextArea: {
		types.InputFieldTypeNone: `<div style="position: relative;">
  <textarea id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}" placeholder="{{ form_print .Loc .Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>{{.Field.Value}}</textarea>
</div>`,
	},
	types.FieldTypeGroup: {
		types.InputFieldTypeNone: `<div style="margin-bottom: 0.5rem; border: 1px solid #dee2e6; border-radius: 0.25rem; background-color: #fff;">
  <div style="padding: 0.5rem 1rem; border-bottom: 1px solid #dee2e6; background-color: #f8f9fa;">
    <h6 style="margin: 0; font-size: 0.875rem; font-weight: 500; color: #212529;" id="{{.Field.Id}}_legend">{{.Field.Legend}}</h6>
  </div>
  <div style="padding: 0.5rem 1rem;" role="group" aria-labelledby="{{.Field.Id}}_legend">
    {{ fields }}
  </div>
</div>`,
	},
	types.FieldTypeError: {
		types.InputFieldTypeNone: `{{range errors}}<div style="display: block; width: 100%; margin-top: 0.25rem; font-size: 0.75rem; color: #dc3545;" role="alert">{{.}}</div>{{end}}`,
	},
	types.FieldTypeLabel: {
		types.InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} style="display: block; margin-bottom: 0.25rem; font-size: 0.875rem; font-weight: 500; color: #212529;" {{with .Field.Id}}id="{{.}}_label"{{end}}>{{ form_print .Loc .Field.Label}}{{ if eq .Field.Required true }}<span style="color: #dc3545;" aria-hidden="true">*</span><span style="position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0;">(required)</span>{{end}}</label>`,
	},
	types.FieldTypeWrapper: {
		types.InputFieldTypeNone: `<div style="margin-bottom: 0.5rem;">
  {{ label }}
  {{ field }}
  {{ range errors }}
  <div style="display: block; width: 100%; margin-top: 0.25rem; font-size: 0.75rem; color: #dc3545;" role="alert">{{ . }}</div>
  {{ end }}
  {{ if .Field.Description }}
  <div style="margin-top: 0.25rem; font-size: 0.75rem; color: #6c757d;" id="{{.Field.Id}}_description">{{.Field.Description}}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeForm: {
		types.InputFieldTypeNone: `<form action="{{.Field.Target}}" method="{{.Field.Method}}" style="max-width: 32rem; margin: 0 auto; padding: 1rem; border: 1px solid #dee2e6; border-radius: 0.25rem; background-color: #fff; box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);" {{ if .Field.Attributes }}{{ form_attributes .Field.Attributes }}{{end}}>
  {{ fields }}
  <div style="margin-top: 1rem; text-align: right;">
    <button type="submit" style="display: inline-block; font-weight: 400; text-align: center; white-space: nowrap; vertical-align: middle; user-select: none; border: 1px solid transparent; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; border-radius: 0.25rem; color: #fff; background-color: #0d6efd; border-color: #0d6efd; cursor: pointer; transition: color 0.15s ease-in-out, background-color 0.15s ease-in-out, border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;">{{ form_print .Loc .Field.Label }}</button>
  </div>
</form>`,
	},
}
