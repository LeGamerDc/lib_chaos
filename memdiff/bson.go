package memdiff

import (
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

var zero = reflect.Value{}

func MarshalRoot(x interface{}) bson.D {
	rv := reflect.Indirect(reflect.ValueOf(x))
	//nolint:govet
	if rv == zero {
		return nil // nil pointer
	}
	vt := rv.Type()
	if vt.Kind() != reflect.Struct {
		return nil
	}
	nf := vt.NumField()
	set := make(bson.D, 0, 8)
	for i := 0; i < nf; i++ {
		f := vt.Field(i)
		if !f.IsExported() { // 1. exclude un-exported field
			continue
		}
		name, needMarshal := formatField(f.Name, f.Tag)
		if !needMarshal { // 2. filter non bson field
			continue
		}
		v := rv.Field(i)
		if v.Kind() != reflect.Pointer {
			if v.CanAddr() {
				v = v.Addr()
			} else {
				fmt.Println("field cant addr", f.Name)
				continue
			}
		}
		if !v.CanInterface() { // 3. filter which cant interface
			continue
		}
		iv := v.Interface()
		if name == "" { // 4. inline field
			set = append(set, MarshalRoot(iv)...)
		} else {
			set = append(set, bson.E{
				Key:   name,
				Value: iv,
			})
		}
	}
	return set
}

func formatField(name string, tag reflect.StructTag) (string, bool) {
	if strings.Contains(tag.Get("memdiff"), "main") {
		return "", false
	}
	b, ok := tag.Lookup("bson")
	if !ok {
		return strings.ToLower(name), true
	}
	if b == "-" {
		return "", false
	}
	if strings.Contains(b, "inline") {
		return "", true
	}
	idx := strings.IndexByte(b, ',')
	if idx == -1 && len(b) > 0 { // `bson:"Name"`
		return b, true
	} else if idx > 0 { // `bson:"Name,omitempty"`
		return b[:idx], true
	}
	return strings.ToLower(name), true // `bson:",omitempty"`
}
