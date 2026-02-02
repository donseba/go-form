package form

import (
	"strings"
	"testing"

	"github.com/donseba/go-form/templates"
	"github.com/donseba/go-form/types"
)

// TestStructWithClassTags tests that class tags are properly processed by the transformer
func TestStructWithClassTags(t *testing.T) {
	type TestForm struct {
		Info         `target:"/submit" method:"post"`
		Email        string `name:"email" form:"input,email" label:"Email Address" required:"true" class:"email-input custom-field"`
		Password     string `name:"password" form:"input,password" label:"Password" required:"true" class:"password-input mb-4"`
		Age          int    `name:"age" form:"input,number" label:"Age" class:"age-input numeric-field"`
		Bio          string `name:"bio" form:"textarea" label:"Biography" rows:"4" class:"bio-textarea large-field"`
		AcceptTerms  bool   `name:"accept_terms" form:"checkbox" label:"Accept Terms" required:"true" class:"terms-checkbox important"`
		Subscription string `name:"subscription" form:"dropdown" label:"Subscription" values:"basic:Basic;premium:Premium;enterprise:Enterprise" class:"subscription-dropdown"`
		Submit       string `form:"input,submit" label:"Submit Form" class:"submit-btn primary-btn"`
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

	// Check each field for correct class value
	expectedClasses := map[string]string{
		"email":        "email-input custom-field",
		"password":     "password-input mb-4",
		"age":          "age-input numeric-field",
		"bio":          "bio-textarea large-field",
		"accept_terms": "terms-checkbox important",
		"subscription": "subscription-dropdown",
		"Submit":       "submit-btn primary-btn",
	}

	// Skip first field (Info)
	for i := 1; i < len(transformer.Fields); i++ {
		field := transformer.Fields[i]
		expectedClass, ok := expectedClasses[field.Name]
		if !ok {
			t.Errorf("Unexpected field name: %s", field.Name)
			continue
		}

		if field.Class != expectedClass {
			t.Errorf("Field %s: expected class '%s', got '%s'", field.Name, expectedClass, field.Class)
		}
	}
}

// TestClassInRenderedHTML tests that class attributes are correctly included in rendered HTML
func TestClassInRenderedHTML(t *testing.T) {
	// Setup basic form with class tag
	type SimpleForm struct {
		Info `target:"/submit" method:"post"`
		Text string `name:"text" form:"input,text" label:"Text Field" class:"text-input custom-class"`
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
		{"TailwindV4", templates.TailwindV4},
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

			// Check if both class values are included in the output
			// Instead of checking for exact attribute format, check that both class values are present
			htmlStr := string(html)
			if !strings.Contains(htmlStr, "text-input") || !strings.Contains(htmlStr, "custom-class") {
				t.Errorf("%s template: class attribute values not found in rendered HTML", ts.name)
				t.Logf("Rendered HTML: %s", htmlStr)
			}
		})
	}
}

// TestGroupWithClass tests that class is properly applied to group fields
func TestGroupWithClass(t *testing.T) {
	type Address struct {
		Street string `name:"street" form:"input" label:"Street" class:"street-input"`
		City   string `name:"city" form:"input" label:"City" class:"city-input"`
	}

	type ContactForm struct {
		Info    `target:"/submit" method:"post"`
		Name    string  `name:"name" form:"input" label:"Name" class:"name-input"`
		Address Address `name:"address" form:"group" label:"Address" legend:"Address Information" class:"address-group"`
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
	if transformer.Fields[1].Class != "name-input" {
		t.Errorf("Name field: expected class 'name-input', got '%s'", transformer.Fields[1].Class)
	}

	if transformer.Fields[2].Class != "address-group" {
		t.Errorf("Address group: expected class 'address-group', got '%s'", transformer.Fields[2].Class)
	}

	// Check nested fields
	addressFields := transformer.Fields[2].Fields
	if len(addressFields) != 2 {
		t.Fatalf("Expected 2 address fields, got %d", len(addressFields))
	}

	if addressFields[0].Class != "street-input" {
		t.Errorf("Street field: expected class 'street-input', got '%s'", addressFields[0].Class)
	}

	if addressFields[1].Class != "city-input" {
		t.Errorf("City field: expected class 'city-input', got '%s'", addressFields[1].Class)
	}
}
