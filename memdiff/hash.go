package memdiff

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"lib_chaos/utils"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	checkRt = reflect.TypeOf(Header{})
	timeRt  = reflect.TypeOf(time.Time{})
	emptyCb = func(string, StructHeader) {}
	zeroMd5 = [md5.Size]byte{}
)

type Header struct {
	Md5 [md5.Size]byte
}

func (c *Header) InitHeader(ptr StructHeader) {
	utils.PanicWrap2(HashThen, ptr, emptyCb)
}

func (c *Header) Updated(m [md5.Size]byte) bool {
	//fmt.Printf("yyy: %x %x\n", m, *c)
	if !bytes.Equal(c.Md5[:], m[:]) {
		c.Md5 = m
		return true
	}
	return false
}

func (c *Header) StructHashChecker() {}

type StructHeader interface {
	Updated([md5.Size]byte) bool
	StructHashChecker()
}

type Callback func(name string, ptr StructHeader)

func HashThen(x StructHeader, cb Callback) {
	hashThen("", x, cb, 0)
}

func hashThen(name string, x StructHeader, cb Callback, d int) {
	rv := reflect.ValueOf(x)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		cb(name, nil)
	} else {
		buf := bytes.NewBuffer(nil)
		process(buf, rv, cb, d)
		m := md5.Sum(buf.Bytes())
		if x.Updated(m) {
			cb(name, x)
		}
	}
}

func Pollute(x StructHeader) {
	pollute(x, 0)
}

func pollute(x StructHeader, d int) {
	rv := reflect.ValueOf(x)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return
		}
		rv = reflect.Indirect(rv)
	}
	x.Updated(zeroMd5)

	if d == 1 || rv.Kind() != reflect.Struct {
		return
	}
	vt := rv.Type()
	nf := vt.NumField()
	for i := 0; i < nf; i++ {
		f := vt.Field(i)
		if strings.Contains(f.Tag.Get("memdiff"), "main") {
			v := rv.Field(i)
			if v.Kind() != reflect.Pointer {
				if v.CanAddr() {
					v = v.Addr()
				} else {
					fmt.Println("field tag memdiff:\"main\" but not implement StructHeader: ", f.Name)
					continue
				}
			}
			if !v.CanInterface() {
				continue
			}
			iv := v.Interface()
			if s, ok := iv.(StructHeader); ok {
				pollute(s, d+1)
			} else {
				fmt.Println("field tag memdiff:\"main\" but not implement StructHeader: ", f.Name)
			}
		}
	}
}

func Serialize(v reflect.Value) string {
	buf := bytes.NewBuffer(nil)
	process(buf, v, emptyCb, 1)
	return buf.String()
}

func process(buf *bytes.Buffer, val reflect.Value, cb Callback, d int) {
	switch val.Kind() {
	case reflect.String:
		buf.WriteByte('"')
		ss := val.String()
		if len(ss) > 32 {
			m5 := md5.Sum([]byte(ss))
			buf.Write(m5[:])
		} else {
			buf.WriteString(val.String())
		}
		buf.WriteByte('"')
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(strconv.FormatInt(val.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		buf.WriteString(strconv.FormatUint(val.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		buf.WriteString(strconv.FormatFloat(val.Float(), 'E', -1, 64))
	case reflect.Bool:
		if val.Bool() {
			buf.WriteByte('t')
		} else {
			buf.WriteByte('f')
		}
	case reflect.Pointer:
		if !val.IsNil() {
			process(buf, reflect.Indirect(val), cb, d)
		} else {
			buf.WriteByte('n')
		}
	case reflect.Array, reflect.Slice:
		buf.WriteByte('[')
		vt := val.Type()
		if vt.Elem().Kind() == reflect.Uint8 {
			m5 := md5.Sum(val.Bytes())
			buf.Write(m5[:])
		} else {
			l := val.Len()
			for i := 0; i < l; i++ {
				if i != 0 {
					buf.WriteByte(',')
				}
				process(buf, val.Index(i), cb, 1)
			}
		}
		buf.WriteByte(']')
	case reflect.Map:
		mk := val.MapKeys()
		items := make([]item, len(mk))
		// Get all values
		for i := range items {
			items[i].name = formatValue(mk[i], cb)
			items[i].value = val.MapIndex(mk[i])
		}

		// Sort values by key
		sort.Sort(itemSorter(items))

		buf.WriteByte('[')
		for i := range items {
			if i != 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(items[i].name)
			buf.WriteByte(':')
			process(buf, items[i].value, cb, 1)
		}
		buf.WriteByte(']')
	case reflect.Struct:
		vt := val.Type()
		if vt == checkRt {
			return
		}
		if vt == timeRt {
			buf.WriteString(formatTime(val))
			return
		}
		nf := vt.NumField()
		buf.WriteByte('{')
		for i := 0; i < nf; i++ {
			f := vt.Field(i)
			if !f.IsExported() { // 1. only care exported field
				continue
			}
			v := val.Field(i)
			if strings.Contains(f.Tag.Get("memdiff"), "main") && d == 0 { // 2. process main entry
				if v.Kind() != reflect.Pointer {
					if v.CanAddr() {
						v = v.Addr()
					} else {
						fmt.Println("field tag memdiff:\"main\" but not implement StructHeader: ", f.Name)
						continue
					}
				}

				if !v.CanInterface() { // 3. exclude unexported field
					continue
				}
				iv := v.Interface()
				if s, ok := iv.(StructHeader); ok {
					name := bsonName(f)
					hashThen(name, s, cb, d+1)
					if v.Kind() == reflect.Pointer && v.IsNil() {
						buf.WriteString(name + ":n")
					}
				} else {
					fmt.Println("field tag memdiff:\"main\" but not implement StructHeader: ", f.Name)
				}
				continue
			}

			if strings.Contains(f.Tag.Get("bson"), "-") { // 4. exclude bson transparent field
				continue
			}

			process(buf, v, cb, d)
		}
		buf.WriteByte('}')
	case reflect.Interface:
		if !val.CanInterface() {
			return
		}
		process(buf, reflect.ValueOf(val.Interface()), cb, d)
	default: // Func Chan UnsafePointer etc
		buf.WriteString(val.Kind().String())
	}
}

type item struct {
	name  string
	value reflect.Value
}

func formatValue(val reflect.Value, cb Callback) string {
	if val.Kind() == reflect.String {
		return "\"" + val.String() + "\""
	}

	var buf bytes.Buffer
	process(&buf, val, cb, 1)

	return buf.String()
}

func formatTime(val reflect.Value) string {
	if val.CanInterface() {
		i := val.Interface()
		t, ok := i.(time.Time)
		if ok {
			return strconv.Itoa(int(t.UnixNano()))
		}
	}
	return "0"
}

type itemSorter []item

func (s itemSorter) Len() int {
	return len(s)
}

func (s itemSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s itemSorter) Less(i, j int) bool {
	return s[i].name < s[j].name
}

func bsonName(f reflect.StructField) string {
	b, ok := f.Tag.Lookup("bson")
	if ok {
		idx := strings.Index(b, ",")
		if idx == -1 && len(b) > 0 {
			return b
		} else if idx > 0 {
			return b[:idx]
		}
	}
	return strings.ToLower(f.Name)
}
