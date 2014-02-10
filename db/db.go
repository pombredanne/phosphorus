package db

import (
	// "github.com/crowdmob/goamz/aws"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/crowdmob/goamz/dynamodb"
	"reflect"
	"strconv"
)

// type IdService interface {
// 	NextId() int64
// }

// type JobState int

// const (
// 	JOB_WAIT JobState = iota
// 	JOB_LOCK
// 	JOB_DEAD
// 	JOB_OK
// )

// type JobDescription struct {
// 	Name     string
// 	Instance func(string) (JobF, error)
// }

// type JobF func() error

// type Job struct {
// 	IndexId  int64    `dynamodb:"_hash"`
// 	Id       int64    `dynamodb:"_range"`
// 	Type     string   `dynamodb:"type",json:"-"`
// 	Argument string   `dynamodb:"argument",json:"-"`
// 	State    JobState `dynamodb:"state",json:"-"`
// 	Fn       *JobF    `json:"-"`
// }

func Dynamize(s interface{}, t *dynamodb.Table) (k *dynamodb.Key, a []dynamodb.Attribute) {
	k = &dynamodb.Key{}
	a = []dynamodb.Attribute{}

	st := reflect.TypeOf(s).Elem()
	sv := reflect.ValueOf(s).Elem()
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		tag := sf.Tag.Get("dynamodb")
		if tag == "" {
			continue
		}
		vf := sv.Field(i)

		if isZero(vf) {
			continue
		}

		switch tag {
		case "_hash":
			k.HashKey = dynamizeKey(vf, t.Key.KeyAttribute.Type)
		case "_range":
			k.RangeKey = dynamizeKey(vf, t.Key.RangeAttribute.Type)
		default:
			a = append(a, *dynamizeAttr(tag, vf))
		}
	}

	return
}

func GetItem(t *dynamodb.Table, s interface{}) error {
	key, _ := Dynamize(s, t)

	attrMap, err := t.GetItem(key)
	if err != nil {
		return err
	}

	return fillAttrs(s, attrMap)
}

func PutItem(t *dynamodb.Table, s interface{}) error {
	key, attrs := Dynamize(s, t)
	_, err := t.PutItem(key.HashKey, key.RangeKey, attrs)
	if err != nil {
		return err
	}
	return nil
}

func ConditionalUpdate(t *dynamodb.Table, update interface{}, expected interface{}) (bool, error) {
	eKey, eAttrs := Dynamize(expected, t)
	_, uAttrs := Dynamize(update, t)

	return t.ConditionalUpdateAttributes(eKey, uAttrs, eAttrs)
}

func AddAttributes(t *dynamodb.Table, s interface{}) error {
	k, a := Dynamize(s, t)
	_, err := t.AddAttributes(k, a)
	return err
}

func b64ify(vf reflect.Value) string {
	buf := &bytes.Buffer{}
	switch vf.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		binary.Write(buf, binary.BigEndian, vf.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		binary.Write(buf, binary.BigEndian, vf.Uint())
	case reflect.String:
		buf.WriteString(vf.String())
	default:
		panic("Can't b64ify unsupported type")
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())

}

func stringify(vf reflect.Value) string {
	switch vf.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(vf.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(vf.Uint(), 10)
	case reflect.String:
		return vf.String()
	default:
		panic("Can't stringify unsupported type")
	}
}

func dynamizeKey(vf reflect.Value, dynamoType string) string {
	switch dynamoType {
	case dynamodb.TYPE_STRING, dynamodb.TYPE_NUMBER:
		return stringify(vf)
	case dynamodb.TYPE_BINARY:
		return b64ify(vf)
	default:
		panic("Can't use set type as key")
	}
}

func dynamizeAttr(name string, vf reflect.Value) *dynamodb.Attribute {
	switch vf.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return dynamodb.NewNumericAttribute(name, stringify(vf))
	case reflect.String:
		return dynamodb.NewStringAttribute(name, stringify(vf))
	case reflect.Slice:
		setVals := make([]string, 0, vf.Len())
		for i := 0; i < vf.Len(); i++ {
			setVals = append(setVals, stringify(vf.Index(i)))
		}

		switch vf.Index(0).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return dynamodb.NewNumericSetAttribute(name, setVals)
		case reflect.String:
			return dynamodb.NewStringSetAttribute(name, setVals)
		default:
			panic("Can't dynamize slices of types other than int, uint, or string")
		}
	default:
		panic("Can't dynamize type other than int, uint, string, or slice")
	}
}

func fillAttrs(s interface{}, attrMap map[string]*dynamodb.Attribute) error {
	st := reflect.TypeOf(s).Elem()
	sv := reflect.ValueOf(s).Elem()

	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		tag := sf.Tag.Get("dynamodb")
		if tag == "" || tag[0] == '_' {
			continue
		}

		attr, exists := attrMap[tag]
		if !exists {
			continue
		}

		vf := sv.Field(i)
		var err error

		switch attr.Type {
		case dynamodb.TYPE_STRING:
			vf.SetString(attr.Value)
		case dynamodb.TYPE_NUMBER:
			err = setNumeric(vf, attr.Value)
		// case dynamodb.TYPE_BINARY:
		// 	err = setBinary(vf, attr.Value)
		case dynamodb.TYPE_STRING_SET:
			err = setStringSet(vf, attr.SetValues)
		case dynamodb.TYPE_NUMBER_SET:
			err = setNumericSet(vf, attr.SetValues)
		// case dynamodb.TYPE_BINARY_SET:
		// 	err = setBinarySet(vf, attr.SetValues)
		default:
			panic("Unrecognized DynamoDB type")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func setNumeric(vf reflect.Value, value string) error {
	switch vf.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		vf.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		vf.SetUint(i)
	default:
		return fmt.Errorf("invalid type!")
	}
	return nil
}

func setStringSet(v reflect.Value, setVals []string) error {
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("invalid type!")
	}

	slice := reflect.MakeSlice(v.Type(), len(setVals), len(setVals))
	for i, setVal := range setVals {
		slice.Index(i).SetString(setVal)
	}
	v.Set(slice)

	return nil
}

func setNumericSet(v reflect.Value, setVals []string) error {
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("invalid type!")
	}

	slice := reflect.MakeSlice(v.Type(), len(setVals), len(setVals))
	for i, setVal := range setVals {
		switch slice.Index(0).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			iv, err := strconv.ParseInt(setVal, 10, 64)
			if err != nil {
				return err
			}
			slice.Index(i).SetInt(iv)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uiv, err := strconv.ParseUint(setVal, 10, 64)
			if err != nil {
				return err
			}
			slice.Index(i).SetUint(uiv)
		case reflect.Float32, reflect.Float64:
			flt, err := strconv.ParseFloat(setVal, 64)
			if err != nil {
				return err
			}
			slice.Index(i).SetFloat(flt)
		default:
			return fmt.Errorf("Invalid type for numeric set")
		}
	}
	v.Set(slice)
	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice:
		return v.Len() == 0
	default:
		return reflect.DeepEqual(reflect.Zero(v.Type()).Interface(), v.Interface())
	}
	return false
}

// func setBinarySet(v reflect.Value, value []string) error {
// 	return fmt.Errorf("eh not really my type")
// }

// func setBinary(v reflect.Value, value string) error {
// 	switch v.Kind() {
// 	case reflect.String:
// 		b, err := base64.StdEncoding.DecodeString(value)
// 		if err != nil {
// 			return err
// 		}
// 		v.SetString(string(b))
// 	}
// 	return fmt.Errorf("eh not really my type")
// }
