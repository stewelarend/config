package config_test

import (
	"testing"

	"github.com/stewelarend/config"
)

func Test1(t *testing.T) {
	values := config.NewValues("test", nil)
	val, ok := values.Get("")
	if ok || val != nil {
		t.Fatalf("get(\"\")->%v,%v", val, ok)
	}

	val, ok = values.Get("a")
	if ok || val != nil {
		t.Fatalf("get(\"a\")->%v,%v", val, ok)
	}

	//we can set a=1 only once and retrieve its value:
	values.Set("a", 1)
	a, ok := values.Get("a")
	if a != 1 || !ok {
		t.Fatalf("get(\"a\")->%v,%v", a, ok)
	}
	if err := values.Set("a", 1); err == nil {
		t.Fatalf("set succeeded after already defined")
	}

	//we can set b="one" only once and retrieve the value:
	values.Set("b", "one")
	b, ok := values.Get("b")
	if b != "one" || !ok {
		t.Fatalf("get(\"b\")->%v,%v", b, ok)
	}
	if err := values.Set("b", "one"); err == nil {
		t.Fatalf("set succeeded after already defined")
	}

	//we can also set an array
	values.Set("c", []string{"one", "two", "three"})
	_c, ok := values.Get("c")
	c, _ := _c.([]string)
	if len(c) != 3 || c[0] != "one" || c[1] != "two" || c[2] != "three" || !ok {
		t.Fatalf("get(\"c\")->%v,%v", c, ok)
	}
	if err := values.Set("c", "four"); err == nil {
		t.Fatalf("set succeeded after already defined")
	}

	//we can also set an object value
	values.Set("d", map[string]interface{}{"one": 1, "two": 2, "three": 3})
	_d, ok := values.Get("d")
	d := _d.(map[string]interface{})
	if len(d) != 3 || d["one"] != 1 || d["two"] != 2 || d["three"] != 3 || !ok {
		t.Fatalf("get(\"d\")->%v,%v", c, ok)
	}
	if err := values.Set("d", "four"); err == nil {
		t.Fatalf("set succeeded after already defined")
	}

	//we can also retrieve items inside an object that we defined, like d.one
	d_one, ok := values.Get("d.one")
	if d_one != 1 || !ok {
		t.Fatalf("get(\"d.one\")->%v,%v", d_one, ok)
	}
	d_two, ok := values.Get("d.two")
	if d_two != 2 || !ok {
		t.Fatalf("get(\"d.two\")->%v,%v", d_one, ok)
	}
	d_three, ok := values.Get("d.three")
	if d_three != 3 || !ok {
		t.Fatalf("get(\"d.three\")->%v,%v", d_one, ok)
	}

	//we cannot set a sub item inside d that already exists to same or new values
	if err := values.Set("d.three", 3); err == nil {
		t.Fatalf("set succeeded after already defined")
	}
	if err := values.Set("d.three", 4); err == nil {
		t.Fatalf("set succeeded after already defined")
	}
	//we can set new sub items though
	if err := values.Set("d.four", 4); err != nil {
		t.Fatalf("failed to set d.four")
	}
	//and retrieve them
	d_four, ok := values.Get("d.four")
	if d_four != 4 || !ok {
		t.Fatalf("get(\"d.four\")->%v,%v", d_one, ok)
	}

	//and retrieve d now consisting of that as well
	_d, ok = values.Get("d")
	d, _ = _d.(map[string]interface{})
	if len(d) != 4 || d["one"] != 1 || d["two"] != 2 || d["three"] != 3 || d["four"] != 4 || !ok {
		t.Fatalf("get(\"d\")->%v,%v", c, ok)
	}
	t.Logf("d=%+v", d)
}
