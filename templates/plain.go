package templates

import "github.com/donseba/go-form/types"

var Plain = map[types.FieldType]map[types.InputFieldType]string{
	types.FieldTypeInput: {
		types.InputFieldTypeText:     `<input type="text" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem; border: 1px solid #ccc; border-radius: 3px; font-size: 0.875rem;">`,
		types.InputFieldTypePassword: `<input type="password" id="{{.Field.Id}}" name="{{.Field.Name}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem; border: 1px solid #ccc; border-radius: 3px; font-size: 0.875rem;">`,
		types.InputFieldTypeEmail:    `<input type="email" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem; border: 1px solid #ccc; border-radius: 3px; font-size: 0.875rem;">`,
		types.InputFieldTypeTel:      `<input type="tel" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem; border: 1px solid #ccc; border-radius: 3px; font-size: 0.875rem;">`,
		types.InputFieldTypeNumber:   `<input type="number" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" step="{{.Field.Step}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem; border: 1px solid #ccc; border-radius: 3px; font-size: 0.875rem;">`,
		types.InputFieldTypeDate:     `<input type="date" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem; border: 1px solid #ccc; border-radius: 3px; font-size: 0.875rem;">`,
		types.InputFieldTypeNone:     `<input id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem; border: 1px solid #ccc; border-radius: 3px; font-size: 0.875rem;">`,
		types.InputFieldTypeHidden:   `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}">`,
	},
	types.FieldTypeCheckbox: {
		types.InputFieldTypeNone: `<div style="margin: 0.25rem 0;">
  <input type="checkbox" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} style="margin-right: 0.375rem;">
  <label for="{{.Field.Id}}" style="cursor: pointer; font-size: 0.875rem;">{{.Field.Label}}</label>
</div>`,
	},
	types.FieldTypeRadios: {
		types.InputFieldTypeNone: `<div style="margin: 0.25rem 0;">
  <input type="radio" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{ if eq .Field.Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} style="margin-right: 0.375rem;">
  <label for="{{.Field.Id}}" style="cursor: pointer; font-size: 0.875rem;">{{.Field.Label}}</label>
</div>`,
	},
	types.FieldTypeDropdown: {
		types.InputFieldTypeNone: `<select id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem; border: 1px solid #ccc; border-radius: 3px; font-size: 0.875rem;">
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeTextArea: {
		types.InputFieldTypeNone: `<textarea id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} style="width: 100%; padding: 0.375rem; border: 1px solid #ccc; border-radius: 3px; font-size: 0.875rem; resize: vertical;">{{.Field.Value}}</textarea>`,
	},
	types.FieldTypeGroup: {
		types.InputFieldTypeNone: `<div style="margin: 0.75rem 0; padding: 0.75rem; border: 1px solid #eee; border-radius: 3px;">
  <h3 style="margin: 0 0 0.75rem 0; font-size: 0.9375rem;">{{.Field.Legend}}</h3>
  {{ fields }}
</div>`,
	},
	types.FieldTypeError: {
		types.InputFieldTypeNone: `{{range errors}}<div style="color: #dc3545; font-size: 0.75rem; margin-top: 0.25rem;">{{.}}</div>{{end}}`,
	},
	types.FieldTypeLabel: {
		types.InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} style="display: block; margin-bottom: 0.375rem; font-size: 0.875rem; font-weight: 500;">{{.Field.Label}}{{ if eq .Field.Required true }}<span style="color: #dc3545;">*</span>{{end}}</label>`,
	},
	types.FieldTypeWrapper: {
		types.InputFieldTypeNone: `<div style="margin-bottom: 0.75rem;">
  {{ label }}
  {{ field }}
  {{ range errors }}
  <div style="color: #dc3545; font-size: 0.75rem; margin-top: 0.25rem;">{{ . }}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeForm: {
		types.InputFieldTypeNone: `<form action="{{.Target}}" method="{{.Method}}" style="max-width: 510px; margin: 0 auto; padding: 2rem; background: #fff; border: 1px solid #e5e7eb; border-radius: 0.75rem; box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05); font-family: Arial, sans-serif; font-size: 0.875rem; line-height: 1.4;">
  {{ fields }}
  <div style="margin-top: 1rem;">
    <button type="submit" style="width: 100%; padding: 0.5rem; background-color: #007bff; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 0.875rem;">Submit</button>
  </div>
</form>`,
	},
}
