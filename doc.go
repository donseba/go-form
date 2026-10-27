// Package form renders HTML forms from Go structs with struct tags, or from a
// Schema built at runtime, through html/template themes. It maps a submission
// back onto the struct, validates it with the rules of the tags and shows the
// errors at their fields. A CSRF middleware and a spam honeypot protect the
// forms.
package form
