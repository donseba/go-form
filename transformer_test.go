package form

import (
	"testing"
)

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

	expected := `{"fields":[{"name":"SomeRandomTextFieldName","value":"","type":"input","inputType":"text","label":"SomeRandomTextFieldName","required":true},{"name":"IntField","value":0,"type":"input","inputType":"number","label":"IntField","step":"1","required":false},{"name":"FloatField","value":0,"type":"input","inputType":"number","label":"FloatField","step":"any","required":false},{"name":"SubGroup","value":{"SubTextField":"","SubIntField":0},"type":"group","label":"SubGroup","required":false,"fields":[{"name":"SubGroup.SubTextField","value":"","type":"input","inputType":"text","label":"SubTextField","required":true},{"name":"SubGroup.SubIntField","value":0,"type":"input","inputType":"number","label":"SubIntField","step":"1","required":false}],"legend":"legendSubGroup"}]}`
	if string(out.JSON()) != expected {
		t.Error("transformer render changed")
	}
}
