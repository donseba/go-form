package templates

import "github.com/donseba/go-form/types"

var BootstrapV5 = map[types.FieldType]map[types.InputFieldType]string{
	types.FieldTypeBase: {
		types.InputFieldTypeNone: `<input type="{{.Type}}" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" placeholder="{{.Field.Placeholder}}" {{if .Field.Required}}required{{end}} class="form-control" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>`,
	},
	types.FieldTypeInput: {
		types.InputFieldTypeNone: `{{ baseInput "Type" "text" "Field" .Field}}`,
		types.InputFieldTypeText: `<div class="position-relative">
  {{ baseInput "Type" "text" "Field" .Field}}
  {{ if .Field.MaxLength }}
  <div class="position-absolute top-0 end-0 mt-2 me-2 small text-muted" id="{{.Field.Id}}_count" aria-hidden="true">0/{{.Field.MaxLength}}</div>
  {{ end }}
</div>`,
		types.InputFieldTypePassword: `<div class="position-relative">
  {{ baseInput "Type" "password" "Field" .Field}}
  <div class="position-absolute top-0 end-0 mt-2 me-2">
    <button type="button" class="btn btn-link btn-sm text-muted p-0" onclick="togglePassword('{{.Field.Id}}')" aria-label="Toggle password visibility">
      <i class="bi bi-eye"></i>
    </button>
  </div>
</div>`,
		types.InputFieldTypeEmail:         `{{ baseInput "Type" "email" "Field" .Field}}`,
		types.InputFieldTypeTel:           `{{ baseInput "Type" "tel" "Field" .Field}}`,
		types.InputFieldTypeNumber:        `<input type="number" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{if .Field.Min}}min="{{.Field.Min}}"{{end}} {{if .Field.Max}}max="{{.Field.Max}}"{{end}} {{if .Field.Step}}step="{{.Field.Step}}"{{end}} {{if .Field.Required}}required{{end}} class="form-control" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>`,
		types.InputFieldTypeDate:          `{{ baseInput "Type" "date" "Field" .Field}}`,
		types.InputFieldTypeDateTimeLocal: `{{ baseInput "Type" "datetime-local" "Field" .Field}}`,
		types.InputFieldTypeTime:          `{{ baseInput "Type" "time" "Field" .Field}}`,
		types.InputFieldTypeWeek:          `{{ baseInput "Type" "week" "Field" .Field}}`,
		types.InputFieldTypeMonth:         `{{ baseInput "Type" "month" "Field" .Field}}`,
		types.InputFieldTypeSearch:        `{{ baseInput "Type" "search" "Field" .Field}}`,
		types.InputFieldTypeUrl:           `{{ baseInput "Type" "url" "Field" .Field}}`,
		types.InputFieldTypeColor: `<div class="d-flex align-items-center gap-2">
  <input type="color" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" class="form-control form-control-color" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  <span class="small text-muted" aria-hidden="true">{{.Field.Value}}</span>
</div>`,
		types.InputFieldTypeRange: `<div class="d-flex align-items-center gap-2">
  <input type="range" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Value}}" {{if .Field.Min}}min="{{.Field.Min}}"{{end}} {{if .Field.Max}}max="{{.Field.Max}}"{{end}} {{if .Field.Step}}step="{{.Field.Step}}"{{end}} class="form-control" oninput="document.getElementById('{{.Field.Id}}_value').textContent = this.value" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  <span id="{{.Field.Id}}_value" class="small text-muted" style="min-width: 3rem; text-align: right;" aria-hidden="true">{{.Field.Value}}</span>
</div>`,
		types.InputFieldTypeImage:  `<input type="image" id="{{.Field.Id}}" name="{{.Field.Name}}" src="{{.Field.Value}}" alt="{{.Field.Label}}" class="form-control" aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>`,
		types.InputFieldTypeSubmit: `<button type="submit" class="btn btn-primary btn-sm" {{ if eq .Field.Disabled true }}disabled{{end}} aria-labelledby="{{.Field.Id}}_label">{{.Field.Label}}</button>`,
		types.InputFieldTypeHidden: `<input type="hidden" name="{{.Field.Name}}" value="{{.Field.Value}}">`,
	},
	types.FieldTypeCheckbox: {
		types.InputFieldTypeNone: `<div class="form-check">
  <input type="checkbox" class="form-check-input" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} {{ if eq .Field.Value true }}checked{{end}} aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  <label class="form-check-label" for="{{.Field.Id}}" id="{{.Field.Id}}_label" style="font-size: 0.875rem; color: #212529; margin-left: 0.25rem;">{{.Field.Label}}</label>
</div>`,
	},
	types.FieldTypeRadios: {
		types.InputFieldTypeNone: `<div role="radiogroup" aria-labelledby="{{.Field.Id}}_label">
  {{ range $k, $option := .Field.Values }}
  <div class="form-check form-check-inline">
    <input type="radio" class="form-check-input" id="{{$.Field.Id}}_{{$k}}" name="{{$.Field.Name}}" value="{{$option.Value}}" {{ if eq $.Field.Value $option.Value }}checked{{end}} {{ if eq $.Field.Required true }}required{{end}} aria-labelledby="{{$.Field.Id}}_{{$k}}_label" {{if $.Field.Description}}aria-describedby="{{$.Field.Id}}_description"{{end}}>
    <label class="form-check-label" for="{{$.Field.Id}}_{{$k}}" id="{{$.Field.Id}}_{{$k}}_label" style="font-size: 0.875rem; color: #212529; margin-left: 0.25rem;">{{$option.Name}}</label>
  </div>
  {{ end }}
</div>`,
		types.InputFieldTypeRadioStruct: `<div class="form-check">
  <input type="radio" class="form-check-input" id="{{.Field.Id}}" name="{{.Field.Name}}" value="{{.Field.Id}}" {{ if eq .Field.Value .Field.Id }}checked{{end}} {{ if eq .Field.Required true }}required{{end}} aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  <label class="form-check-label" for="{{.Field.Id}}" id="{{.Field.Id}}_label" style="font-size: 0.875rem; color: #212529; margin-left: 0.25rem;">{{.Field.Label}}</label>
</div>`,
	},
	types.FieldTypeDropdown: {
		types.InputFieldTypeNone: `<select class="form-select form-select-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeDropdownMapped: {
		types.InputFieldTypeNone: `<select class="form-select form-select-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" {{ if eq .Field.Required true }}required{{end}} aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>
  {{ $value := .Field.Value }}
  {{ range $k, $option := .Field.Values }}
  <option value="{{$option.Value}}" {{ if eq $value $option.Value }}selected{{ end }} {{ if eq $option.Disabled true }}disabled{{ end }}>{{$option.Name}}</option>
  {{ end }}
</select>`,
	},
	types.FieldTypeTextArea: {
		types.InputFieldTypeNone: `<div class="position-relative">
  <textarea class="form-control form-control-sm" id="{{.Field.Id}}" name="{{.Field.Name}}" rows="{{.Field.Rows}}" cols="{{.Field.Cols}}" placeholder="{{.Field.Placeholder}}" {{ if eq .Field.Required true }}required{{end}} aria-labelledby="{{.Field.Id}}_label" {{if .Field.Description}}aria-describedby="{{.Field.Id}}_description"{{end}}>{{.Field.Value}}</textarea>
  {{ if .Field.MaxLength }}
  <div class="position-absolute bottom-0 end-0 mb-2 me-2 small text-muted" id="{{.Field.Id}}_count" aria-hidden="true">0/{{.Field.MaxLength}}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeGroup: {
		types.InputFieldTypeNone: `<div class="card card-sm mb-2">
  <div class="card-header py-1">
    <h6 class="card-title mb-0" id="{{.Field.Id}}_legend">{{.Field.Legend}}</h6>
  </div>
  <div class="card-body py-2" role="group" aria-labelledby="{{.Field.Id}}_legend">
    {{ fields }}
  </div>
</div>`,
	},
	types.FieldTypeError: {
		types.InputFieldTypeNone: `{{range errors}}<div class="invalid-feedback d-block small" role="alert">{{.}}</div>{{end}}`,
	},
	types.FieldTypeLabel: {
		types.InputFieldTypeNone: `<label {{with .Field.Id}}for="{{.}}"{{end}} class="form-label small mb-1" {{with .Field.Id}}id="{{.}}_label"{{end}}>{{.Field.Label}}{{ if eq .Field.Required true }}<span class="text-danger" aria-hidden="true">*</span><span class="visually-hidden">(required)</span>{{end}}</label>`,
	},
	types.FieldTypeWrapper: {
		types.InputFieldTypeNone: `<div class="mb-2">
  {{ label }}
  {{ field }}
  {{ range errors }}
  <div class="invalid-feedback d-block small" role="alert">{{ . }}</div>
  {{ end }}
  {{ if .Field.Description }}
  <div class="form-text small" id="{{.Field.Id}}_description">{{.Field.Description}}</div>
  {{ end }}
</div>`,
	},
	types.FieldTypeForm: {
		types.InputFieldTypeNone: `<form action="{{.Field.Target}}" method="{{.Field.Method}}" class="mx-auto border rounded shadow-sm p-3" style="max-width: 32rem;" novalidate>
  {{ fields }}
  <div class="d-grid gap-2 mt-3">
    <button type="submit" class="btn btn-primary btn-sm" >Submit</button>
  </div>
</form>`,
	},
}
