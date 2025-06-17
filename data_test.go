package form

import (
	"strings"
	"testing"

	"github.com/donseba/go-form/templates"
	"github.com/donseba/go-form/types"
)

// TestStructWithDataTags tests that data tags are properly processed by the transformer
func TestStructWithDataTags(t *testing.T) {
	type TestForm struct {
		Info         `target:"/submit" method:"post"`
		Email        string `name:"email" form:"input,email" label:"Email Address" required:"true" data:"validate=email,mask=true"`
		Password     string `name:"password" form:"input,password" label:"Password" required:"true" data:"min-length=8,password-meter=true"`
		Age          int    `name:"age" form:"input,number" label:"Age" data:"min-value=18,validation=numeric"`
		Bio          string `name:"bio" form:"textarea" label:"Biography" rows:"4" data:"max-length=500,counter=true"`
		AcceptTerms  bool   `name:"accept_terms" form:"checkbox" label:"Accept Terms" required:"true" data:"group=terms,required-msg=You must accept the terms"`
		Subscription string `name:"subscription" form:"dropdown" label:"Subscription" values:"basic:Basic;premium:Premium;enterprise:Enterprise" data:"change-handler=updatePrice"`
		Submit       string `form:"input,submit" label:"Submit Form" data:"analytics=form-submit,action=register"`
	}

	testForm := TestForm{
		Email:        "test@example.com",
		Password:     "secret123",
		Age:          30,
		Bio:          "This is my biography",
		AcceptTerms:  true,
		Subscription: "premium",
	}

	// Create transformer
	transformer, err := NewTransformer(testForm)
	if err != nil {
		t.Fatalf("Failed to create transformer: %v", err)
	}

	// Check each field for correct data attributes
	expectedData := map[string]map[string]string{
		"email": {
			"validate": "email",
			"mask":     "true",
		},
		"password": {
			"min-length":     "8",
			"password-meter": "true",
		},
		"age": {
			"min-value":  "18",
			"validation": "numeric",
		},
		"bio": {
			"max-length": "500",
			"counter":    "true",
		},
		"accept_terms": {
			"group":        "terms",
			"required-msg": "You must accept the terms",
		},
		"subscription": {
			"change-handler": "updatePrice",
		},
		"Submit": {
			"analytics": "form-submit",
			"action":    "register",
		},
	}

	// Skip first field (Info)
	for i := 1; i < len(transformer.Fields); i++ {
		field := transformer.Fields[i]
		expectedFieldData, ok := expectedData[field.Name]
		if !ok {
			t.Errorf("Unexpected field name: %s", field.Name)
			continue
		}

		// Check if Data map exists
		if field.Data == nil {
			t.Errorf("Field %s: Data map is nil", field.Name)
			continue
		}

		// Check each expected data attribute
		for key, expectedValue := range expectedFieldData {
			actualValue, exists := field.Data[key]
			if !exists {
				t.Errorf("Field %s: missing data attribute '%s'", field.Name, key)
				continue
			}
			if actualValue != expectedValue {
				t.Errorf("Field %s: expected data attribute '%s' to be '%s', got '%s'", field.Name, key, expectedValue, actualValue)
			}
		}

		// Check if there are any unexpected data attributes
		for key := range field.Data {
			if _, exists := expectedFieldData[key]; !exists {
				t.Errorf("Field %s: unexpected data attribute '%s'", field.Name, key)
			}
		}
	}
}

// TestDataInRenderedHTML tests that data attributes are correctly included in rendered HTML
func TestDataInRenderedHTML(t *testing.T) {
	// Setup basic form with data tag
	type SimpleForm struct {
		Info `target:"/submit" method:"post"`
		Text string `name:"text" form:"input,text" label:"Text Field" data:"validate=required,max-length=100,toggle=other-field"`
	}

	simpleForm := SimpleForm{
		Text: "Sample text",
	}

	// Test with different templates
	templateSets := []struct {
		name     string
		template types.TemplateMap
	}{
		{"Bootstrap", templates.BootstrapV5},
		{"Plain", templates.Plain},
		{"Tailwind", templates.TailwindV3},
	}

	for _, ts := range templateSets {
		t.Run(ts.name, func(t *testing.T) {
			// Create a Form with the template
			form := NewForm(ts.template)

			// Render the form using the existing formRenderFunc method
			html, err := form.formRenderFunc(&DefaultLocalizer{}, simpleForm, nil)
			if err != nil {
				t.Fatalf("Failed to render form: %v", err)
			}

			// Check if data attributes are included in the output
			htmlStr := string(html)
			dataAttrs := []string{"data-validate=\"required\"", "data-max-length=\"100\"", "data-toggle=\"other-field\""}

			for _, attr := range dataAttrs {
				if !strings.Contains(htmlStr, attr) {
					t.Errorf("%s template: data attribute %q not found in rendered HTML", ts.name, attr)
					t.Logf("Rendered HTML: %s", htmlStr)
				}
			}
		})
	}
}

// TestNestedDataAttributes tests handling of data attributes in nested structures
func TestNestedDataAttributes(t *testing.T) {
	type Address struct {
		Street string `name:"street" form:"input" label:"Street" data:"autocomplete=street-address"`
		City   string `name:"city" form:"input" label:"City" data:"autocomplete=address-level2"`
	}

	type ContactForm struct {
		Info    `target:"/submit" method:"post"`
		Name    string  `name:"name" form:"input" label:"Name" data:"autocomplete=name"`
		Address Address `name:"address" form:"group" label:"Address" legend:"Address Information" data:"section=address"`
	}

	contactForm := ContactForm{
		Name: "John Doe",
		Address: Address{
			Street: "123 Main St",
			City:   "Anytown",
		},
	}

	transformer, err := NewTransformer(contactForm)
	if err != nil {
		t.Fatalf("Failed to create transformer: %v", err)
	}

	// Check main fields
	if transformer.Fields[1].Data == nil || transformer.Fields[1].Data["autocomplete"] != "name" {
		t.Errorf("Name field: expected data attribute 'autocomplete=name', got %v", transformer.Fields[1].Data)
	}

	if transformer.Fields[2].Data == nil || transformer.Fields[2].Data["section"] != "address" {
		t.Errorf("Address group: expected data attribute 'section=address', got %v", transformer.Fields[2].Data)
	}

	// Check nested fields
	addressFields := transformer.Fields[2].Fields
	if len(addressFields) != 2 {
		t.Fatalf("Expected 2 address fields, got %d", len(addressFields))
	}

	if addressFields[0].Data == nil || addressFields[0].Data["autocomplete"] != "street-address" {
		t.Errorf("Street field: expected data attribute 'autocomplete=street-address', got %v", addressFields[0].Data)
	}

	if addressFields[1].Data == nil || addressFields[1].Data["autocomplete"] != "address-level2" {
		t.Errorf("City field: expected data attribute 'autocomplete=address-level2', got %v", addressFields[1].Data)
	}
}

// TestComplexDataAttributeValues tests handling of data attributes with complex values
func TestComplexDataAttributeValues(t *testing.T) {
	type ComplexForm struct {
		Info  `target:"/submit" method:"post"`
		Field string `name:"complex" form:"input" label:"Complex Field" data:"json-config={\"key\":\"value\",\"items\":[1,2,3]},message=This is a message with spaces"`
	}

	form := ComplexForm{
		Field: "Test",
	}

	transformer, err := NewTransformer(form)
	if err != nil {
		t.Fatalf("Failed to create transformer: %v", err)
	}

	// Check complex data attributes
	field := transformer.Fields[1]
	expected := map[string]string{
		"json-config": `{"key":"value","items":[1,2,3]}`,
		"message":     "This is a message with spaces",
	}

	if field.Data == nil {
		t.Fatal("Field data is nil")
	}

	for key, expectedValue := range expected {
		if field.Data[key] != expectedValue {
			t.Errorf("Field %s: expected data attribute %q to be %q, got %q",
				field.Name, key, expectedValue, field.Data[key])
		}
	}

	// Test rendering to ensure complex values are properly escaped in HTML
	form1 := NewForm(templates.Plain)
	html, err := form1.formRenderFunc(&DefaultLocalizer{}, form, nil)
	if err != nil {
		t.Fatalf("Failed to render form: %v", err)
	}

	htmlStr := string(html)
	if !strings.Contains(htmlStr, "data-json-config=\"{&#34;key&#34;:&#34;value&#34;,&#34;items&#34;:[1,2,3]}\"") {
		t.Errorf("JSON data not properly rendered and escaped in HTML output")
		t.Logf("Rendered HTML: %s", htmlStr)
	}

	if !strings.Contains(htmlStr, "data-message=\"This is a message with spaces\"") {
		t.Errorf("Message with spaces not properly rendered in HTML output")
		t.Logf("Rendered HTML: %s", htmlStr)
	}
}
