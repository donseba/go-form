package templates

import "github.com/donseba/go-form/types"

var BootstrapV5 = map[types.FieldType]map[types.InputFieldType]string{
	types.FieldTypeInput: {
		types.InputFieldTypeText:     `<input type="text" class="form-control form-control-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypePassword: `<input type="password" class="form-control form-control-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeEmail:    `<input type="email" class="form-control form-control-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeTel:      `<input type="tel" class="form-control form-control-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeNumber:   `<input type="number" class="form-control form-control-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" step="{{.Field.Step}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeDate:     `<input type="date" class="form-control form-control-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeNone:     `<input class="form-control form-control-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>`,
		types.InputFieldTypeHidden:   `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}">`,
	},
	types.FieldTypeCheckbox: {
		types.InputFieldTypeNone: `<div class="form-check form-check-inline">
  <input type="checkbox" class="form-check-input" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}}>
</div>`,
	},
	types.FieldTypeRadios: {
		types.InputFieldTypeNone: `<br/>{{ range $k, $option := .Field.Values }}
<div class="form-check form-check-inline">
  <input type="radio" class="form-check-input" id="{{$.Field.Id}}_{{$k}}" name="{{$.Field.Name}}" value="{{$option.Value}}" {{ if eq $.Field.Value $option.Value }}checked{{end}} {{ if eq $.Field.Required true }}required{{end}}>
  <label class="form-check-label" for="{{$.Field.Id}}_{{$k}}">{{$option.Name}}</label>
</div>
{{ end }}`,
		types.InputFieldTypeRadioStruct: `<div class="form-check form-check-inline">
  <input type="radio" class="form-check-input" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Id}}" {{ if eq .Field.Value .Field.Id }}checked{{end}} {{ if eq .Field.Required true }}required{{end}}>
</div>`,
	},
	types.FieldTypeDropdown: {
		types.InputFieldTypeNone: `<select class="form-select form-select-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeTextArea: {
		types.InputFieldTypeNone: `<textarea class="form-control form-control-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}}>{{.Field.Value}}</textarea>`,
	},
	types.FieldTypeGroup: {
		types.InputFieldTypeNone: `<div class="card card-sm mb-2">
  <div class="card-header py-1">
    <h6 class="card-title mb-0">{{.Field.Legend}}</h6>
  </div>
  <div class="card-body py-2">
    {{ fields }}
  </div>
</div>`,
	},
	types.FieldTypeError: {
		types.InputFieldTypeNone: `{{range errors}}<div class="invalid-feedback d-block small">{{.}}</div>{{end}}`,
	},
	types.FieldTypeLabel: {
		types.InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} class="form-label small mb-1">{{.Field.Label}}{{ if eq .Field.Required true }}<span class="text-danger">*</span>{{end}}</label>`,
	},
	types.FieldTypeWrapper: {
		types.InputFieldTypeNone: `<div class="mb-2">
  {{ label }}
  {{ field }}
  {{ range errors }}
  <div class="invalid-feedback d-block small">{{ . }}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeForm: {
		types.InputFieldTypeNone: `<form action="{{.Field.Target}}" method="{{.Field.Method}}" class="mx-auto border rounded shadow-sm p-3" style="max-width: 32rem;">
  {{ fields }}
  <div class="d-grid gap-2 mt-3">
    <button type="submit" class="btn btn-primary btn-sm">Submit</button>
  </div>
</form>`,
	},
}
