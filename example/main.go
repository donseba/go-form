package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/donseba/go-form"
	"github.com/donseba/go-form/templates"
)

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type (
	Gender      string
	ExampleForm struct {
		form.Info
		ID int `form:"input,hidden"`
		// Basic input fields
		Username string `form:"input,text" label:"Username" placeholder:"Enter your username" required:"true"`
		Password string `form:"input,password" label:"Password" placeholder:"Enter your password" required:"true"`
		Email    string `form:"input,email" label:"Email" placeholder:"Enter your email" required:"true"`
		Phone    string `form:"input,tel" label:"Phone label" placeholder:"Enter your phone number"`
		Age      int    `form:"input,number" label:"Age" placeholder:"Enter your age" step:"1"`
		Birthday string `form:"input,date" label:"Birthday" placeholder:"Enter your birthday"`

		// Checkbox and radio fields
		Active      bool   `form:"checkbox" label:"Active"`
		RadioOption string `form:"radios" label:"Radio Option" values:"option1:Option 1;option2:Option 2;option3:Option 3"`

		// Color field
		Color string `form:"input,color" label:"Favorite Color" placeholder:"Select your favorite color"`

		// Range field
		RangeValue int `form:"input,range" label:"Range Value" min:"25" max:"75" step:"1"`

		//date and time fields
		DateTimeLocal string `form:"input,datetime-local" label:"Date and Time" placeholder:"Select date and time"`
		Time          string `form:"input,time" label:"Time" placeholder:"Select time"`
		Week          string `form:"input,week" label:"Week" placeholder:"Select week"`
		Month         string `form:"input,month" label:"Month" placeholder:"Select month"`

		// enum field
		Gender Gender `label:"Gender"`

		// Dropdown fields
		Country string `form:"dropdown" label:"Country" values:"us:United States;ca:Canada;uk:United Kingdom"`
		Message string `form:"textarea" label:"Message" placeholder:"Enter your message" rows:"5" cols:"50"`
		Hidden  string `form:"input,hidden" value:"hidden value"`

		RadioGroup RadioGroupBlock `form:"radios" label:"Radio Group"`

		// Nested group
		Address Address `legend:"Address Information"`
	}

	RadioGroupBlock struct {
		Option1 bool `name:"RadioGroup" label:"first option"`
		Option2 bool `name:"RadioGroup" label:"second option"`
	}

	Address struct {
		Street1 string `form:"input,text" label:"Street Address" placeholder:"Enter street address" required:"true"`
		City    string `form:"input,text" label:"City" placeholder:"Enter city" required:"true"`
		State   string `form:"input,text" label:"State" placeholder:"Enter state" required:"true"`
		Zip     string `form:"input,text" label:"ZIP Code" placeholder:"Enter ZIP code" required:"true"`
	}
)

func (i Gender) String() string {
	return string(i)
}
func (i Gender) Enum() []any {
	return []interface{}{GenderMale, GenderFemale, GenderOther}
}

func main() {
	formData := ExampleForm{
		Info: form.Info{
			Target: "/submit",
			Method: "POST",
		},
		ID:       1,
		Username: "john.doe",
		Email:    "john.doe@example.com",
		Phone:    "123-456-7890",
		Age:      30,
		Birthday: "1990-01-01",
		Active:   true,
		Gender:   "male",
		Country:  "us",
		Message:  "Hello, this is a test message!",
		Hidden:   "secret-value",
		RadioGroup: RadioGroupBlock{
			Option1: true,
			Option2: false,
		},
		Address: Address{
			Street1: "123 Main St",
			City:    "New York",
			State:   "NY",
			Zip:     "10001",
		},
	}

	// Create the template
	tmpl := template.Must(template.New("example").
		Funcs(map[string]any{
			"form_render": func(formInstance form.Form, errors []form.FieldError) template.HTML {
				return template.HTML("")
			},
		}).
		Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Form Examples</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        .template-section {
            margin: 2rem 0;
            padding: 1rem;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        .template-title {
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid #eee;
        }
        .nav-links {
            margin-bottom: 2rem;
        }
        .nav-links a {
            margin-right: 1rem;
            padding: 0.5rem 1rem;
            text-decoration: none;
            color: #333;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        .nav-links a:hover {
            background-color: #f5f5f5;
        }
    </style>
</head>
<body>
    <div class="container py-4">
        <h1 class="mb-4">Form Template Examples</h1>
        
        <div class="nav-links">
            <a href="/plain">Plain Template</a>
            <a href="/bootstrap">Bootstrap 5 Template</a>
            <a href="/tailwind">Tailwind CSS Template</a>
        </div>

        {{ form_render .Form .Errors }}
    </div>
</body>
</html>
`))

	// Create HTTP handler for each template type
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/plain", http.StatusFound)
	})

	http.HandleFunc("/plain", func(w http.ResponseWriter, r *http.Request) {
		formInstance := form.NewForm(templates.Plain)

		renderPage(w, formInstance, formData, tmpl)
	})

	http.HandleFunc("/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		formInstance := form.NewForm(templates.BootstrapV5)

		renderPage(w, formInstance, formData, tmpl)
	})

	http.HandleFunc("/tailwind", func(w http.ResponseWriter, r *http.Request) {
		formInstance := form.NewForm(templates.TailwindV3)

		renderPage(w, formInstance, formData, tmpl)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func renderPage(w http.ResponseWriter, formInstance form.Form, formData ExampleForm, tmpl *template.Template) {
	data := struct {
		Form   ExampleForm
		Errors []form.FieldError
	}{
		Form: formData,
		Errors: []form.FieldError{
			fieldError{
				Field: "Email",
				Issue: "is already taken",
			},
			fieldError{
				Field: "Address.Street1",
				Issue: "is required",
			},
		},
	}

	tmplCopy, err := tmpl.Clone()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmplCopy.Funcs(formInstance.FuncMap())

	if err := tmplCopy.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type fieldError struct {
	Field string
	Issue string
}

func (fe fieldError) Error() string {
	return fmt.Sprintf("%s: %s", fe.Field, fe.Issue)
}

func (fe fieldError) FieldError() (field, err string) {
	return fe.Field, fe.Issue
}
