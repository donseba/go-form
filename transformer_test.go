package form

import "testing"

type ModelA struct {
	TextField  string `required:"true" name:"SomeRandomTextFieldName"`
	IntField   int64
	FloatField float64
	SubGroup   struct {
		SubTextField string `required:"true"`
		SubIntField  int64
	} `legend:"legendSubGroup"`
}

func TestNewTransformer(t *testing.T) {
	out, err := NewTransformer(&ModelA{})
	if err != nil {
		t.Error(err)
	}

	expected := `{"fields":[{"type":"","name":"SomeRandomTextFieldName","id":"SomeRandomTextFieldName","label":"SomeRandomTextFieldName","value":"","required":true},{"type":"input","inputType":"number","name":"IntField","id":"IntField","label":"IntField","value":0,"step":"1"},{"type":"input","inputType":"number","name":"FloatField","id":"FloatField","label":"FloatField","value":0,"step":"any"},{"type":"group","name":"SubGroup","id":"SubGroup","label":"SubGroup","value":{"SubTextField":"","SubIntField":0},"legend":"legendSubGroup","fields":[{"type":"","name":"SubGroup.SubTextField","id":"SubGroup.SubTextField","label":"SubTextField","value":"","required":true},{"type":"input","inputType":"number","name":"SubGroup.SubIntField","id":"SubGroup.SubIntField","label":"SubIntField","value":0,"step":"1"}]}]}`
	if string(out.JSON()) != expected {
		t.Error("transformer render changed")
	}
}
