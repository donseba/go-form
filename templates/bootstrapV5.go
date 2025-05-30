package templates

import "github.com/donseba/go-form/types"

var BootstrapV5 = map[types.FieldType]map[types.InputFieldType]string{
	types.FieldTypeInput: {
		types.InputFieldTypeText:     `<input type="text" class="form-control" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypePassword: `<input type="password" class="form-control" id="{{.Field.Id}}" name="{{.Field.Name}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeEmail:    `<input type="email" class="form-control" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeTel:      `<input type="tel" class="form-control" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeNumber:   `<input type="number" class="form-control" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" step="{{.Field.Step}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeDate:     `<input type="date" class="form-control" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeNone:     `<input class="form-control" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeHidden:   `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}">`,
	},
	types.FieldTypeCheckbox: {
		types.InputFieldTypeNone: `<div class="form-check">
  <input type="checkbox" class="form-check-input" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}}>
  <label class="form-check-label" for="{{.Field.Id}}">{{.Field.Label}}</label>
</div>`,
	},
	types.FieldTypeRadios: {
		types.InputFieldTypeNone: `<div class="form-check">
  <input type="radio" class="form-check-input" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{ if eq .Field.Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}}>
  <label class="form-check-label" for="{{.Field.Id}}">{{.Field.Label}}</label>
</div>`,
	},
	types.FieldTypeDropdown: {
		types.InputFieldTypeNone: `<select class="form-select" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeTextArea: {
		types.InputFieldTypeNone: `<textarea class="form-control" id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>{{.Field.Value}}</textarea>`,
	},
	types.FieldTypeGroup: {
		types.InputFieldTypeNone: `<div class="card mb-3">
  <div class="card-header">
    <h5 class="card-title mb-0">{{.Field.Legend}}</h5>
  </div>
  <div class="card-body">
    {{ fields }}
  </div>
</div>`,
	},
	types.FieldTypeError: {
		types.InputFieldTypeNone: `{{range errors}}<div class="invalid-feedback d-block">{{.}}</div>{{end}}`,
	},
	types.FieldTypeLabel: {
		types.InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} class="form-label">{{.Field.Label}}{{ if eq .Field.Required true }}<span class="text-danger">*</span>{{end}}</label>`,
	},
	types.FieldTypeWrapper: {
		types.InputFieldTypeNone: `<div class="mb-3">
  {{ label }}
  {{ field }}
  {{ range errors }}
  <div class="invalid-feedback d-block">{{ . }}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeForm: {
		types.InputFieldTypeNone: `<form action="{{.Target}}" method="{{.Method}}" class="mx-auto max-w-lg space-y-6 rounded-xl border border-gray-200 bg-white p-8 shadow-sm dark:border-gray-700 dark:bg-gray-700 dark:shadow-md">
  {{ fields }}
  <div class="d-grid gap-2">
    <button type="submit" class="btn btn-primary">Submit</button>
  </div>
</form>`,
	},
}
