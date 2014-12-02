// Package sflag is the only known flags package variant, at the time of writing, that is 100% DRY, free of fugly pointer syntax and uses clean struct syntax.
// Implementation makes use of reflection and struct tags.
//
// BUG() Presence of a boolean flag requires that there be no STANDALONE true or false parameters, use "--Foo=true" syntax instead of "--Foo true".
package sflag

import (
	"flag"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// Parse runs through the struct members and the struct tags of the struct that will hold the program commandline options.
// It uses that information to set up the call to golang's flag package's Parse() function
func Parse(ss interface{}) {
	if reflect.TypeOf(ss).Kind() != reflect.Ptr {
		panic("sflag.Parse was not provided a pointer arg")
	}
	sstype := reflect.TypeOf(ss).Elem()
	ssvalue := reflect.ValueOf(ss).Elem()

	if sstype.Kind() != reflect.Struct {
		panic("sflag.Parse was not provided a pointer to a struct")
	}

	moreusage := ""

	var flags = *flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	var argsptr *[]string

	hasBoolArg := false

	for ii := 0; ii < sstype.NumField(); ii++ {
		pp := sstype.Field(ii)
		vv := ssvalue.Field(ii)
		if pp.Anonymous {
			continue
		}
		if pp.Name == "Usage" {
			continue
		}
		if pp.Type.String() == "[]string" {
			up := unsafe.Pointer(vv.UnsafeAddr())
			argsptr = (*[]string)(up)
			continue
		}
		tag := strings.TrimSpace((string)(pp.Tag))
		if tag == "" {
			continue
		}
		splitChar := tag[0:1]
		if strings.Contains("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", splitChar) {
			splitChar = "|"
		} else {
			tag = tag[1:]
		}
		parts := strings.Split(tag, splitChar)
		part0 := ""
		part1 := ""
		if len(parts) > 0 {
			part0 = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			part1 = strings.TrimSpace(parts[1])
		}
		if len(parts) > 0 {
			moreusage += "\n\t--" + pp.Name + ": " + part1 + " <-- Default, " + pp.Type.String() + " # " + part0
		}
		if len(parts) == 1 {
			up := unsafe.Pointer(vv.UnsafeAddr())
			switch pp.Type.String() {
			case "string":
				flags.StringVar((*string)(up), pp.Name, vv.String(), " <--default, string # "+part0)
			case "int":
				flags.IntVar((*int)(up), pp.Name, int(vv.Int()), " <--default, int # "+part0)
			case "bool":
				flags.BoolVar((*bool)(up), pp.Name, bool(vv.Bool()), " <--default, bool # "+part0)
				hasBoolArg = true
			case "int64":
				flags.Int64Var((*int64)(up), pp.Name, vv.Int(), " <--default, int64 # "+part0)
			case "float64":
				flags.Float64Var((*float64)(up), pp.Name, vv.Float(), " <--default, float64 # "+part0)
			}
		}
		if len(parts) == 2 {
			up := unsafe.Pointer(vv.UnsafeAddr())
			switch pp.Type.String() {
			case "string":
				vv.SetString(part1)
				flags.StringVar((*string)(up), pp.Name, part1, " <--default, string # "+part0)
			case "int":
				inum, _ := strconv.ParseInt(part1, 10, 64)
				vv.SetInt(inum)
				flags.IntVar((*int)(up), pp.Name, int(inum), " <--default, int # "+part0)
			case "bool":
				bnum, _ := strconv.ParseBool(part1)
				vv.SetBool(bnum)
				flags.BoolVar((*bool)(up), pp.Name, bool(bnum), " <--default, bool # "+part0)
				hasBoolArg = true
			case "int64":
				jnum, _ := strconv.ParseInt(part1, 10, 64)
				vv.SetInt(jnum)
				flags.Int64Var((*int64)(up), pp.Name, jnum, " <--default, int64 # "+part0)
			case "float64":
				fnum, _ := strconv.ParseFloat(part1, 64)
				vv.SetFloat(fnum)
				flags.Float64Var((*float64)(up), pp.Name, fnum, " <--default, float64 # "+part0)
			}
		}
	}

	pp, _ := sstype.FieldByName("Usage")
	vv := ssvalue.FieldByName("Usage")
	vv.SetString("\n Usage of " + os.Args[0] + " # " + (string)(pp.Tag) + "\n ARGS:" + moreusage)
	if hasBoolArg {
		for _, arg := range os.Args[1:] {
			switch strings.ToLower(arg) {
			case "true", "false":
				panic("Golang flag package requires \"--Foo=bar\" instead of \"--Foo bar\" syntax for bool args")
			}
		}
	}
	flags.Parse(os.Args[1:])
	if argsptr != nil {
		*argsptr = make([]string, len(flags.Args()))
		copy(*argsptr, flags.Args())
	}
}
