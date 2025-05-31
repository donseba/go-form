package templates

import "github.com/donseba/go-form/types"

var Plain = map[types.FieldType]map[types.InputFieldType]string{
	types.FieldTypeInput: {
		types.InputFieldTypeText:     `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="text" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} style="width: 100%; padding: 0.25rem 0.5rem; font-size: 0.875rem; border: 1px solid #ced4da; border-radius: 0.25rem;">`,
		types.InputFieldTypePassword: `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="password" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} style="width: 100%; padding: 0.25rem 0.5rem; font-size: 0.875rem; border: 1px solid #ced4da; border-radius: 0.25rem;">`,
		types.InputFieldTypeEmail:    `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="email" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} style="width: 100%; padding: 0.25rem 0.5rem; font-size: 0.875rem; border: 1px solid #ced4da; border-radius: 0.25rem;">`,
		types.InputFieldTypeTel:      `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="tel" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} style="width: 100%; padding: 0.25rem 0.5rem; font-size: 0.875rem; border: 1px solid #ced4da; border-radius: 0.25rem;">`,
		types.InputFieldTypeNumber:   `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="number" step="{{.Field.Step}}" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} style="width: 100%; padding: 0.25rem 0.5rem; font-size: 0.875rem; border: 1px solid #ced4da; border-radius: 0.25rem;">`,
		types.InputFieldTypeDate:     `<input id="{{.Field.Id}}" name="{{.Field.Name}}" type="date" placeholder="{{.Field.Placeholder}}" {{with .Field.Value}}value="{{.}}"{{end}} {{ if eq .Required true }}required{{end}} style="width: 100%; padding: 0.25rem 0.5rem; font-size: 0.875rem; border: 1px solid #ced4da; border-radius: 0.25rem;">`,
		types.InputFieldTypeNone:     `<input id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.25rem 0.5rem; font-size: 0.875rem; border: 1px solid #ced4da; border-radius: 0.25rem;">`,
		types.InputFieldTypeHidden:   `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}" />`,
	},
	types.FieldTypeCheckbox: {
		types.InputFieldTypeNone: `<div style="display: inline-flex; align-items: center;">
  <input id="{{.Field.Id}}" name="{{.Field.Name}}" type="checkbox" {{ if eq .Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} style="margin: 0;">
</div>`,
	},
	types.FieldTypeRadios: {
		types.InputFieldTypeNone: `{{ range $k, $option := .Field.Values }}
<div style="display: inline-flex; align-items: center; margin-right: 1rem;">
  <input id="{{$.Field.Id}}_{{$k}}" name="{{$.Field.Name}}" type="radio" value="{{$option.Value}}" {{ if eq $.Field.Value $option.Value }}checked{{end}} {{ if eq $.Field.Required true }}required{{end}} style="margin: 0;">
  <label for="{{$.Field.Id}}_{{$k}}" style="margin-left: 0.5rem; font-size: 0.875rem;">{{$option.Name}}</label>
</div>
{{ end }}`,
		types.InputFieldTypeRadioStruct: `<div style="display: inline-flex; align-items: center;">
  <input id="{{.Field.Id}}" name="{{.Field.Name}}" type="radio" value="{{.Field.Id}}" {{ if eq .Field.Value .Field.Id }}checked{{end}} {{ if eq .Field.Required true }}required{{end}} style="margin: 0;">
</div>`,
	},
	types.FieldTypeDropdown: {
		types.InputFieldTypeNone: `<select {{with .Field.Id}}id="{{.}}"{{end}} name="{{.Field.Name}}" style="width: 100%; padding: 0.25rem 0.5rem; font-size: 0.875rem; border: 1px solid #ced4da; border-radius: 0.25rem;">
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeTextArea: {
		types.InputFieldTypeNone: `<textarea id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.25rem 0.5rem; font-size: 0.875rem; border: 1px solid #ced4da; border-radius: 0.25rem;">{{.Field.Value}}</textarea>`,
	},
	types.FieldTypeGroup: {
		types.InputFieldTypeNone: `<div style="border: 1px solid #dee2e6; border-radius: 0.25rem; padding: 0.5rem; margin-bottom: 0.5rem;">
  <h6 style="margin: 0 0 0.5rem 0; font-size: 0.875rem; font-weight: 500;">{{.Field.Legend}}</h6>
  <div style="display: flex; flex-direction: column; gap: 0.5rem;">
    {{ fields }}
  </div>
</div>`,
	},
	types.FieldTypeError: {
		types.InputFieldTypeNone: `{{range errors}}<div style="margin-top: 0.25rem; font-size: 0.75rem; color: #dc3545;">{{.}}</div>{{end}}`,
	},
	types.FieldTypeLabel: {
		types.InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} style="display: block; margin-bottom: 0.25rem; font-size: 0.875rem; font-weight: 500;">{{.Field.Label}}{{ if eq .Field.Required true }}<span style="color: #dc3545;">*</span>{{end}}</label>`,
	},
	types.FieldTypeWrapper: {
		types.InputFieldTypeNone: `<div style="margin-bottom: 0.5rem;">
  {{ label }}
  {{ field }}
  {{ range errors }}
  <div style="margin-top: 0.25rem; font-size: 0.75rem; color: #dc3545;">{{ . }}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeForm: {
		types.InputFieldTypeNone: `<form action="{{.Field.Target}}" method="{{.Field.Method}}" style="max-width: 32rem; margin: 0 auto; border: 1px solid #dee2e6; border-radius: 0.25rem; padding: 0.75rem; box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);">
  {{ fields }}
  <div style="margin-top: 0.75rem;">
    <button type="submit" style="width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; font-weight: 500; color: #fff; background-color: #0d6efd; border: 1px solid #0d6efd; border-radius: 0.25rem; cursor: pointer;">Submit</button>
  </div>
</form>`,
	},
}
