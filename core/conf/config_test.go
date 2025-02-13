package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/fs"
	"github.com/zeromicro/go-zero/core/hash"
)

var dupErr dupKeyError

func TestLoadConfig_notExists(t *testing.T) {
	assert.NotNil(t, Load("not_a_file", nil))
}

func TestLoadConfig_notRecogFile(t *testing.T) {
	filename, err := fs.TempFilenameWithText("hello")
	assert.Nil(t, err)
	defer os.Remove(filename)
	assert.NotNil(t, LoadConfig(filename, nil))
}

func TestConfigJson(t *testing.T) {
	tests := []string{
		".json",
		".yaml",
		".yml",
	}
	text := `{
	"a": "foo",
	"b": 1,
	"c": "${FOO}",
	"d": "abcd!@#$112"
}`
	for _, test := range tests {
		test := test
		t.Run(test, func(t *testing.T) {
			os.Setenv("FOO", "2")
			defer os.Unsetenv("FOO")
			tmpfile, err := createTempFile(test, text)
			assert.Nil(t, err)
			defer os.Remove(tmpfile)

			var val struct {
				A string `json:"a"`
				B int    `json:"b"`
				C string `json:"c"`
				D string `json:"d"`
			}
			MustLoad(tmpfile, &val)
			assert.Equal(t, "foo", val.A)
			assert.Equal(t, 1, val.B)
			assert.Equal(t, "${FOO}", val.C)
			assert.Equal(t, "abcd!@#$112", val.D)
		})
	}
}

func TestLoadFromJsonBytesArray(t *testing.T) {
	input := []byte(`{"users": [{"name": "foo"}, {"Name": "bar"}]}`)
	var val struct {
		Users []struct {
			Name string
		}
	}

	assert.NoError(t, LoadConfigFromJsonBytes(input, &val))
	var expect []string
	for _, user := range val.Users {
		expect = append(expect, user.Name)
	}
	assert.EqualValues(t, []string{"foo", "bar"}, expect)
}

func TestConfigToml(t *testing.T) {
	text := `a = "foo"
b = 1
c = "${FOO}"
d = "abcd!@#$112"
`
	os.Setenv("FOO", "2")
	defer os.Unsetenv("FOO")
	tmpfile, err := createTempFile(".toml", text)
	assert.Nil(t, err)
	defer os.Remove(tmpfile)

	var val struct {
		A string `json:"a"`
		B int    `json:"b"`
		C string `json:"c"`
		D string `json:"d"`
	}
	MustLoad(tmpfile, &val)
	assert.Equal(t, "foo", val.A)
	assert.Equal(t, 1, val.B)
	assert.Equal(t, "${FOO}", val.C)
	assert.Equal(t, "abcd!@#$112", val.D)
}

func TestConfigOptional(t *testing.T) {
	text := `a = "foo"
b = 1
c = "FOO"
d = "abcd"
`
	tmpfile, err := createTempFile(".toml", text)
	assert.Nil(t, err)
	defer os.Remove(tmpfile)

	var val struct {
		A string `json:"a"`
		B int    `json:"b,optional"`
		C string `json:"c,optional=B"`
		D string `json:"d,optional=b"`
	}
	if assert.NoError(t, Load(tmpfile, &val)) {
		assert.Equal(t, "foo", val.A)
		assert.Equal(t, 1, val.B)
		assert.Equal(t, "FOO", val.C)
		assert.Equal(t, "abcd", val.D)
	}
}

func TestConfigJsonCanonical(t *testing.T) {
	text := []byte(`{"a": "foo", "B": "bar"}`)

	var val1 struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	var val2 struct {
		A string
		B string
	}
	assert.NoError(t, LoadFromJsonBytes(text, &val1))
	assert.Equal(t, "foo", val1.A)
	assert.Equal(t, "bar", val1.B)
	assert.NoError(t, LoadFromJsonBytes(text, &val2))
	assert.Equal(t, "foo", val2.A)
	assert.Equal(t, "bar", val2.B)
}

func TestConfigTomlCanonical(t *testing.T) {
	text := []byte(`a = "foo"
B = "bar"`)

	var val1 struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	var val2 struct {
		A string
		B string
	}
	assert.NoError(t, LoadFromTomlBytes(text, &val1))
	assert.Equal(t, "foo", val1.A)
	assert.Equal(t, "bar", val1.B)
	assert.NoError(t, LoadFromTomlBytes(text, &val2))
	assert.Equal(t, "foo", val2.A)
	assert.Equal(t, "bar", val2.B)
}

func TestConfigYamlCanonical(t *testing.T) {
	text := []byte(`a: foo
B: bar`)

	var val1 struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	var val2 struct {
		A string
		B string
	}
	assert.NoError(t, LoadConfigFromYamlBytes(text, &val1))
	assert.Equal(t, "foo", val1.A)
	assert.Equal(t, "bar", val1.B)
	assert.NoError(t, LoadFromYamlBytes(text, &val2))
	assert.Equal(t, "foo", val2.A)
	assert.Equal(t, "bar", val2.B)
}

func TestConfigTomlEnv(t *testing.T) {
	text := `a = "foo"
b = 1
c = "${FOO}"
d = "abcd!@#112"
`
	os.Setenv("FOO", "2")
	defer os.Unsetenv("FOO")
	tmpfile, err := createTempFile(".toml", text)
	assert.Nil(t, err)
	defer os.Remove(tmpfile)

	var val struct {
		A string `json:"a"`
		B int    `json:"b"`
		C string `json:"c"`
		D string `json:"d"`
	}

	MustLoad(tmpfile, &val, UseEnv())
	assert.Equal(t, "foo", val.A)
	assert.Equal(t, 1, val.B)
	assert.Equal(t, "2", val.C)
	assert.Equal(t, "abcd!@#112", val.D)
}

func TestConfigJsonEnv(t *testing.T) {
	tests := []string{
		".json",
		".yaml",
		".yml",
	}
	text := `{
	"a": "foo",
	"b": 1,
	"c": "${FOO}",
	"d": "abcd!@#$a12 3"
}`
	for _, test := range tests {
		test := test
		t.Run(test, func(t *testing.T) {
			os.Setenv("FOO", "2")
			defer os.Unsetenv("FOO")
			tmpfile, err := createTempFile(test, text)
			assert.Nil(t, err)
			defer os.Remove(tmpfile)

			var val struct {
				A string `json:"a"`
				B int    `json:"b"`
				C string `json:"c"`
				D string `json:"d"`
			}
			MustLoad(tmpfile, &val, UseEnv())
			assert.Equal(t, "foo", val.A)
			assert.Equal(t, 1, val.B)
			assert.Equal(t, "2", val.C)
			assert.Equal(t, "abcd!@# 3", val.D)
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  "",
			expect: "",
		},
		{
			input:  "A",
			expect: "a",
		},
		{
			input:  "a",
			expect: "a",
		},
		{
			input:  "hello_world",
			expect: "hello_world",
		},
		{
			input:  "Hello_world",
			expect: "hello_world",
		},
		{
			input:  "hello_World",
			expect: "hello_world",
		},
		{
			input:  "helloWorld",
			expect: "helloworld",
		},
		{
			input:  "HelloWorld",
			expect: "helloworld",
		},
		{
			input:  "hello World",
			expect: "hello world",
		},
		{
			input:  "Hello World",
			expect: "hello world",
		},
		{
			input:  "Hello World",
			expect: "hello world",
		},
		{
			input:  "Hello World foo_bar",
			expect: "hello world foo_bar",
		},
		{
			input:  "Hello World foo_Bar",
			expect: "hello world foo_bar",
		},
		{
			input:  "Hello World Foo_bar",
			expect: "hello world foo_bar",
		},
		{
			input:  "Hello World Foo_Bar",
			expect: "hello world foo_bar",
		},
		{
			input:  "Hello.World Foo_Bar",
			expect: "hello.world foo_bar",
		},
		{
			input:  "你好 World Foo_Bar",
			expect: "你好 world foo_bar",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			assert.Equal(t, test.expect, toLowerCase(test.input))
		})
	}
}

func TestLoadFromJsonBytesError(t *testing.T) {
	var val struct{}
	assert.Error(t, LoadFromJsonBytes([]byte(`hello`), &val))
}

func TestLoadFromTomlBytesError(t *testing.T) {
	var val struct{}
	assert.Error(t, LoadFromTomlBytes([]byte(`hello`), &val))
}

func TestLoadFromYamlBytesError(t *testing.T) {
	var val struct{}
	assert.Error(t, LoadFromYamlBytes([]byte(`':hello`), &val))
}

func TestLoadFromYamlBytes(t *testing.T) {
	input := []byte(`layer1:
  layer2:
    layer3: foo`)
	var val struct {
		Layer1 struct {
			Layer2 struct {
				Layer3 string
			}
		}
	}

	assert.NoError(t, LoadFromYamlBytes(input, &val))
	assert.Equal(t, "foo", val.Layer1.Layer2.Layer3)
}

func TestLoadFromYamlBytesTerm(t *testing.T) {
	input := []byte(`layer1:
  layer2:
    tls_conf: foo`)
	var val struct {
		Layer1 struct {
			Layer2 struct {
				Layer3 string `json:"tls_conf"`
			}
		}
	}

	assert.NoError(t, LoadFromYamlBytes(input, &val))
	assert.Equal(t, "foo", val.Layer1.Layer2.Layer3)
}

func TestLoadFromYamlBytesLayers(t *testing.T) {
	input := []byte(`layer1:
  layer2:
    layer3: foo`)
	var val struct {
		Value string `json:"Layer1.Layer2.Layer3"`
	}

	assert.NoError(t, LoadFromYamlBytes(input, &val))
	assert.Equal(t, "foo", val.Value)
}

func TestLoadFromYamlItemOverlay(t *testing.T) {
	type (
		Redis struct {
			Host string
			Port int
		}

		RedisKey struct {
			Redis
			Key string
		}

		Server struct {
			Redis RedisKey
		}

		TestConfig struct {
			Server
			Redis Redis
		}
	)

	input := []byte(`Redis:
  Host: localhost
  Port: 6379
  Key: test
`)

	var c TestConfig
	assert.ErrorAs(t, LoadFromYamlBytes(input, &c), &dupErr)
}

func TestLoadFromYamlItemOverlayReverse(t *testing.T) {
	type (
		Redis struct {
			Host string
			Port int
		}

		RedisKey struct {
			Redis
			Key string
		}

		Server struct {
			Redis Redis
		}

		TestConfig struct {
			Redis RedisKey
			Server
		}
	)

	input := []byte(`Redis:
  Host: localhost
  Port: 6379
  Key: test
`)

	var c TestConfig
	assert.ErrorAs(t, LoadFromYamlBytes(input, &c), &dupErr)
}

func TestLoadFromYamlItemOverlayWithMap(t *testing.T) {
	type (
		Redis struct {
			Host string
			Port int
		}

		RedisKey struct {
			Redis
			Key string
		}

		Server struct {
			Redis RedisKey
		}

		TestConfig struct {
			Server
			Redis map[string]interface{}
		}
	)

	input := []byte(`Redis:
  Host: localhost
  Port: 6379
  Key: test
`)

	var c TestConfig
	assert.ErrorAs(t, LoadFromYamlBytes(input, &c), &dupErr)
}

func TestUnmarshalJsonBytesMap(t *testing.T) {
	input := []byte(`{"foo":{"/mtproto.RPCTos": "bff.bff","bar":"baz"}}`)

	var val struct {
		Foo map[string]string
	}

	assert.NoError(t, LoadFromJsonBytes(input, &val))
	assert.Equal(t, "bff.bff", val.Foo["/mtproto.RPCTos"])
	assert.Equal(t, "baz", val.Foo["bar"])
}

func TestUnmarshalJsonBytesMapWithSliceElements(t *testing.T) {
	input := []byte(`{"foo":{"/mtproto.RPCTos": ["bff.bff", "any"],"bar":["baz", "qux"]}}`)

	var val struct {
		Foo map[string][]string
	}

	assert.NoError(t, LoadFromJsonBytes(input, &val))
	assert.EqualValues(t, []string{"bff.bff", "any"}, val.Foo["/mtproto.RPCTos"])
	assert.EqualValues(t, []string{"baz", "qux"}, val.Foo["bar"])
}

func TestUnmarshalJsonBytesMapWithSliceOfStructs(t *testing.T) {
	input := []byte(`{"foo":{
	"/mtproto.RPCTos": [{"bar": "any"}],
	"bar":[{"bar": "qux"}, {"bar": "ever"}]}}`)

	var val struct {
		Foo map[string][]struct {
			Bar string
		}
	}

	assert.NoError(t, LoadFromJsonBytes(input, &val))
	assert.Equal(t, 1, len(val.Foo["/mtproto.RPCTos"]))
	assert.Equal(t, "any", val.Foo["/mtproto.RPCTos"][0].Bar)
	assert.Equal(t, 2, len(val.Foo["bar"]))
	assert.Equal(t, "qux", val.Foo["bar"][0].Bar)
	assert.Equal(t, "ever", val.Foo["bar"][1].Bar)
}

func TestUnmarshalJsonBytesWithAnonymousField(t *testing.T) {
	type (
		Int int

		InnerConf struct {
			Name string
		}

		Conf struct {
			Int
			InnerConf
		}
	)

	var (
		input = []byte(`{"Name": "hello", "int": 3}`)
		c     Conf
	)
	assert.NoError(t, LoadFromJsonBytes(input, &c))
	assert.Equal(t, "hello", c.Name)
	assert.Equal(t, Int(3), c.Int)
}

func TestUnmarshalJsonBytesWithMapValueOfStruct(t *testing.T) {
	type (
		Value struct {
			Name string
		}

		Config struct {
			Items map[string]Value
		}
	)

	var inputs = [][]byte{
		[]byte(`{"Items": {"Key":{"Name": "foo"}}}`),
		[]byte(`{"Items": {"Key":{"Name": "foo"}}}`),
		[]byte(`{"items": {"key":{"name": "foo"}}}`),
		[]byte(`{"items": {"key":{"name": "foo"}}}`),
	}
	for _, input := range inputs {
		var c Config
		if assert.NoError(t, LoadFromJsonBytes(input, &c)) {
			assert.Equal(t, 1, len(c.Items))
			for _, v := range c.Items {
				assert.Equal(t, "foo", v.Name)
			}
		}
	}
}

func TestUnmarshalJsonBytesWithMapTypeValueOfStruct(t *testing.T) {
	type (
		Value struct {
			Name string
		}

		Map map[string]Value

		Config struct {
			Map
		}
	)

	var inputs = [][]byte{
		[]byte(`{"Map": {"Key":{"Name": "foo"}}}`),
		[]byte(`{"Map": {"Key":{"Name": "foo"}}}`),
		[]byte(`{"map": {"key":{"name": "foo"}}}`),
		[]byte(`{"map": {"key":{"name": "foo"}}}`),
	}
	for _, input := range inputs {
		var c Config
		if assert.NoError(t, LoadFromJsonBytes(input, &c)) {
			assert.Equal(t, 1, len(c.Map))
			for _, v := range c.Map {
				assert.Equal(t, "foo", v.Name)
			}
		}
	}
}

func Test_FieldOverwrite(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		type Base struct {
			Name string
		}

		type St1 struct {
			Base
			Name2 string
		}

		type St2 struct {
			Base
			Name2 string
		}

		type St3 struct {
			*Base
			Name2 string
		}

		type St4 struct {
			*Base
			Name2 *string
		}

		validate := func(val any) {
			input := []byte(`{"Name": "hello", "Name2": "world"}`)
			assert.NoError(t, LoadFromJsonBytes(input, val))
		}

		validate(&St1{})
		validate(&St2{})
		validate(&St3{})
		validate(&St4{})
	})

	t.Run("Inherit Override", func(t *testing.T) {
		type Base struct {
			Name string
		}

		type St1 struct {
			Base
			Name string
		}

		type St2 struct {
			Base
			Name int
		}

		type St3 struct {
			*Base
			Name int
		}

		type St4 struct {
			*Base
			Name *string
		}

		validate := func(val any) {
			input := []byte(`{"Name": "hello"}`)
			err := LoadFromJsonBytes(input, val)
			assert.ErrorAs(t, err, &dupErr)
			assert.Equal(t, newDupKeyError("name").Error(), err.Error())
		}

		validate(&St1{})
		validate(&St2{})
		validate(&St3{})
		validate(&St4{})
	})

	t.Run("Inherit more", func(t *testing.T) {
		type Base1 struct {
			Name string
		}

		type St0 struct {
			Base1
			Name string
		}

		type St1 struct {
			St0
			Name string
		}

		type St2 struct {
			St0
			Name int
		}

		type St3 struct {
			*St0
			Name int
		}

		type St4 struct {
			*St0
			Name *int
		}

		validate := func(val any) {
			input := []byte(`{"Name": "hello"}`)
			err := LoadFromJsonBytes(input, val)
			assert.ErrorAs(t, err, &dupErr)
			assert.Equal(t, newDupKeyError("name").Error(), err.Error())
		}

		validate(&St0{})
		validate(&St1{})
		validate(&St2{})
		validate(&St3{})
		validate(&St4{})
	})
}

func TestFieldOverwriteComplicated(t *testing.T) {
	t.Run("double maps", func(t *testing.T) {
		type (
			Base1 struct {
				Values map[string]string
			}
			Base2 struct {
				Values map[string]string
			}
			Config struct {
				Base1
				Base2
			}
		)

		var c Config
		input := []byte(`{"Values": {"Key": "Value"}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("merge children", func(t *testing.T) {
		type (
			Inner1 struct {
				Name string
			}
			Inner2 struct {
				Age int
			}
			Base1 struct {
				Inner Inner1
			}
			Base2 struct {
				Inner Inner2
			}
			Config struct {
				Base1
				Base2
			}
		)

		var c Config
		input := []byte(`{"Inner": {"Name": "foo", "Age": 10}}`)
		if assert.NoError(t, LoadFromJsonBytes(input, &c)) {
			assert.Equal(t, "foo", c.Base1.Inner.Name)
			assert.Equal(t, 10, c.Base2.Inner.Age)
		}
	})

	t.Run("overwritten maps", func(t *testing.T) {
		type (
			Inner struct {
				Map map[string]string
			}
			Config struct {
				Map map[string]string
				Inner
			}
		)

		var c Config
		input := []byte(`{"Inner": {"Map": {"Key": "Value"}}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten nested maps", func(t *testing.T) {
		type (
			Inner struct {
				Map map[string]string
			}
			Middle1 struct {
				Map map[string]string
				Inner
			}
			Middle2 struct {
				Map map[string]string
				Inner
			}
			Config struct {
				Middle1
				Middle2
			}
		)

		var c Config
		input := []byte(`{"Middle1": {"Inner": {"Map": {"Key": "Value"}}}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten outer/inner maps", func(t *testing.T) {
		type (
			Inner struct {
				Map map[string]string
			}
			Middle struct {
				Inner
				Map map[string]string
			}
			Config struct {
				Middle
			}
		)

		var c Config
		input := []byte(`{"Middle": {"Inner": {"Map": {"Key": "Value"}}}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten anonymous maps", func(t *testing.T) {
		type (
			Inner struct {
				Map map[string]string
			}
			Middle struct {
				Inner
				Map map[string]string
			}
			Elem   map[string]Middle
			Config struct {
				Elem
			}
		)

		var c Config
		input := []byte(`{"Elem": {"Key": {"Inner": {"Map": {"Key": "Value"}}}}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten primitive and map", func(t *testing.T) {
		type (
			Inner struct {
				Value string
			}
			Elem  map[string]Inner
			Named struct {
				Elem string
			}
			Config struct {
				Named
				Elem
			}
		)

		var c Config
		input := []byte(`{"Elem": {"Key": {"Value": "Value"}}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten map and slice", func(t *testing.T) {
		type (
			Inner struct {
				Value string
			}
			Elem  []Inner
			Named struct {
				Elem string
			}
			Config struct {
				Named
				Elem
			}
		)

		var c Config
		input := []byte(`{"Elem": {"Key": {"Value": "Value"}}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten map and string", func(t *testing.T) {
		type (
			Elem  string
			Named struct {
				Elem string
			}
			Config struct {
				Named
				Elem
			}
		)

		var c Config
		input := []byte(`{"Elem": {"Key": {"Value": "Value"}}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})
}

func TestLoadNamedFieldOverwritten(t *testing.T) {
	t.Run("overwritten named struct", func(t *testing.T) {
		type (
			Elem  string
			Named struct {
				Elem string
			}
			Base struct {
				Named
				Elem
			}
			Config struct {
				Val Base
			}
		)

		var c Config
		input := []byte(`{"Val": {"Elem": {"Key": {"Value": "Value"}}}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten named []struct", func(t *testing.T) {
		type (
			Elem  string
			Named struct {
				Elem string
			}
			Base struct {
				Named
				Elem
			}
			Config struct {
				Vals []Base
			}
		)

		var c Config
		input := []byte(`{"Vals": [{"Elem": {"Key": {"Value": "Value"}}}]}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten named map[string]struct", func(t *testing.T) {
		type (
			Elem  string
			Named struct {
				Elem string
			}
			Base struct {
				Named
				Elem
			}
			Config struct {
				Vals map[string]Base
			}
		)

		var c Config
		input := []byte(`{"Vals": {"Key": {"Elem": {"Key": {"Value": "Value"}}}}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten named *struct", func(t *testing.T) {
		type (
			Elem  string
			Named struct {
				Elem string
			}
			Base struct {
				Named
				Elem
			}
			Config struct {
				Vals *Base
			}
		)

		var c Config
		input := []byte(`{"Vals": [{"Elem": {"Key": {"Value": "Value"}}}]}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten named struct", func(t *testing.T) {
		type (
			Named struct {
				Elem string
			}
			Base struct {
				Named
				Elem Named
			}
			Config struct {
				Val Base
			}
		)

		var c Config
		input := []byte(`{"Val": {"Elem": "Value"}}`)
		assert.ErrorAs(t, LoadFromJsonBytes(input, &c), &dupErr)
	})

	t.Run("overwritten named struct", func(t *testing.T) {
		type Config struct {
			Val chan int
		}

		var c Config
		input := []byte(`{"Val": 1}`)
		assert.Error(t, LoadFromJsonBytes(input, &c))
	})
}

func createTempFile(ext, text string) (string, error) {
	tmpfile, err := os.CreateTemp(os.TempDir(), hash.Md5Hex([]byte(text))+"*"+ext)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(tmpfile.Name(), []byte(text), os.ModeTemporary); err != nil {
		return "", err
	}

	filename := tmpfile.Name()
	if err = tmpfile.Close(); err != nil {
		return "", err
	}

	return filename, nil
}

func TestFillDefaultUnmarshal(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		type St struct{}
		err := FillDefault(St{})
		assert.Error(t, err)
	})

	t.Run("not nil", func(t *testing.T) {
		type St struct{}
		err := FillDefault(&St{})
		assert.NoError(t, err)
	})

	t.Run("default", func(t *testing.T) {
		type St struct {
			A string `json:",default=a"`
			B string
		}
		var st St
		err := FillDefault(&st)
		assert.NoError(t, err)
		assert.Equal(t, st.A, "a")
	})

	t.Run("env", func(t *testing.T) {
		type St struct {
			A string `json:",default=a"`
			B string
			C string `json:",env=TEST_C"`
		}
		t.Setenv("TEST_C", "c")

		var st St
		err := FillDefault(&st)
		assert.NoError(t, err)
		assert.Equal(t, st.A, "a")
		assert.Equal(t, st.C, "c")
	})

	t.Run("has vaue", func(t *testing.T) {
		type St struct {
			A string `json:",default=a"`
			B string
		}
		var st = St{
			A: "b",
		}
		err := FillDefault(&st)
		assert.Error(t, err)
	})
}
