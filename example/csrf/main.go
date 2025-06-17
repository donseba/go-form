package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/donseba/go-form"
	"github.com/donseba/go-form/templates"
)

// LoginForm demonstrates a form with CSRF protection
type LoginForm struct {
	form.Info
	Email      string `form:"input,email" label:"Email" required:"true" data:"autocomplete=email"`
	Password   string `form:"input,password" label:"Password" required:"true" minLength:"10" data:"autocomplete=current-password"`
	RememberMe bool   `form:"checkbox" label:"Remember Me"`
}

func main() {
	// Create a form renderer with CSRF support
	formRenderer := form.NewForm(templates.BootstrapV5)

	// Create a router (we'll use the default ServeMux for simplicity)
	mux := http.NewServeMux()

	// Create our form handler
	formHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create the login form
		loginForm := LoginForm{
			Info: form.Info{
				Target:     "/login",
				Method:     "post",
				SubmitText: "Log In",
			},
		}

		// Inject the CSRF token from the request context
		form.InjectCSRFToken(r, &loginForm.Info)

		// Process form submission
		var errs form.FieldErrors

		if r.Method == http.MethodPost {
			// Map form data to struct
			if err := form.MapForm(r, &loginForm); err != nil {
				http.Error(w, "Error mapping form: "+err.Error(), http.StatusBadRequest)
				return
			}

			// Validate the form data (the CSRF token is already validated by the middleware)
			errs = formRenderer.ValidateForm(&loginForm)
		}

		// Create HTML template with form_render function
		tmpl := template.Must(template.New("index").Funcs(formRenderer.FuncMap()).Parse(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>CSRF Protection Example</title>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/css/bootstrap.min.css" rel="stylesheet">
				<style>
					body { padding: 2rem; }
					.container { max-width: 600px; }
				</style>
			</head>
			<body>
				<div class="container">
					<h1 class="mb-4">CSRF Protection Example</h1>
					
					<div style="margin-bottom:1em">
						<a href="http://localhost:8080/">Templates Example</a> |
						<a href="http://localhost:8081/">Validation Example</a> |
						<a href="http://localhost:8082/">Translation Example</a>
					</div>
					
					<div class="mb-4">
						<div class="alert alert-info">
							This example demonstrates how to add CSRF protection to your forms using middleware.
							The CSRF token is automatically added as a hidden field and validated on submission.
						</div>
					</div>
	
					{{ form_render .Form .Errors }}
	
					<div class="mt-5">
						<h3>How it works:</h3>
						<ol>
							<li>Middleware generates a random CSRF token when rendering the form</li>
							<li>Token is stored in the session and included in the form as a hidden field</li>
							<li>When the form is submitted, middleware validates the token before processing</li>
							<li>After validation, a new token is generated for the next request</li>
						</ol>
					</div>
				</div>
			</body>
			</html>
		`))

		// Execute the template with the form data
		err := tmpl.Execute(w, map[string]any{
			"Form":   loginForm,
			"Errors": errs,
		})

		if err != nil {
			http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		}
	})

	// Apply the CSRF middleware to our handler
	protectedHandler := formRenderer.CSRFMiddleware()(formHandler)

	// Register the handler with the router
	mux.Handle("/", protectedHandler)

	// Start the server
	fmt.Println("Server running at http://localhost:8083/")
	log.Fatal(http.ListenAndServe(":8083", mux))
}
