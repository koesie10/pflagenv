package pflagenv

import (
	"errors"
	"fmt"
	"github.com/fatih/camelcase"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"reflect"
	"strings"
	"time"
)

func init() {
	viper.AllowEmptyEnv(true)
}

var valueType = reflect.TypeOf((*pflag.Value)(nil)).Elem()

func Setup(fset *pflag.FlagSet, c interface{}) error {
	return setup(fset, c, "", "")
}

func setup(fset *pflag.FlagSet, c interface{}, baseEnv, baseFlag string) error {
	val := reflect.ValueOf(c)
	if val.Kind() != reflect.Ptr {
		return errors.New("result must be a pointer")
	}

	val = val.Elem()
	if !val.CanAddr() {
		return errors.New("result must be addressable (a pointer)")
	}

	if val.Kind() != reflect.Struct {
		return errors.New("result must be a struct")
	}

	structType := val.Type()

	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		v := val.Field(i)

		// Unexported fields should not be set
		if !v.CanSet() {
			continue
		}

		env := f.Tag.Get("env")
		flag := f.Tag.Get("flag")
		shorthand := ""
		desc := f.Tag.Get("desc")

		if strings.ContainsRune(flag, ',') {
			parts := strings.Split(flag, ",")
			flag = parts[0]
			shorthand = parts[1]
		}

		if env != "" && strings.ContainsRune(env, ',') {
			parts := strings.Split(env, ",")
			env = parts[0]
		}

		if env == "" {
			env = strings.ToUpper(strings.Join(camelcase.Split(f.Name), "_"))
		}
		if flag == "" {
			flag = strings.ToLower(strings.Join(camelcase.Split(f.Name), "-"))
		}

		// This doesn't work properly with mapstructure and environment variables
		/*if baseEnv != "" {
			env = baseEnv + "_" + env
		}*/
		if baseFlag != "" {
			flag = baseFlag + "-" + flag
		}

		desc += fmt.Sprintf(" (environment %s)", env)

		p := v.Addr()

		if p.Type().Implements(valueType) {
			fset.VarP(p.Interface().(pflag.Value), flag, shorthand, desc)
		} else {
			switch f.Type.Kind() {
			case reflect.String:
				fset.StringVarP(p.Interface().(*string), flag, shorthand, v.String(), desc)
			case reflect.Int:
				fset.IntVarP(p.Interface().(*int), flag, shorthand, int(v.Int()), desc)
			case reflect.Int8:
				fset.Int8VarP(p.Interface().(*int8), flag, shorthand, int8(v.Int()), desc)
			case reflect.Int16:
				fset.Int16VarP(p.Interface().(*int16), flag, shorthand, int16(v.Int()), desc)
			case reflect.Int32:
				fset.Int32VarP(p.Interface().(*int32), flag, shorthand, int32(v.Int()), desc)
			case reflect.Int64:
				if f.Type == reflect.TypeOf(time.Duration(0)) {
					fset.DurationVarP(p.Interface().(*time.Duration), flag, shorthand, v.Interface().(time.Duration), desc)
					break
				}

				fset.Int64VarP(p.Interface().(*int64), flag, shorthand, v.Int(), desc)
			case reflect.Uint:
				fset.UintVarP(p.Interface().(*uint), flag, shorthand, uint(v.Uint()), desc)
			case reflect.Uint8:
				fset.Uint8VarP(p.Interface().(*uint8), flag, shorthand, uint8(v.Int()), desc)
			case reflect.Uint16:
				fset.Uint16VarP(p.Interface().(*uint16), flag, shorthand, uint16(v.Int()), desc)
			case reflect.Uint32:
				fset.Uint32VarP(p.Interface().(*uint32), flag, shorthand, uint32(v.Int()), desc)
			case reflect.Uint64:
				fset.Uint64VarP(p.Interface().(*uint64), flag, shorthand, v.Uint(), desc)
			case reflect.Float64:
				fset.Float64VarP(p.Interface().(*float64), flag, shorthand, v.Float(), desc)
			case reflect.Bool:
				fset.BoolVarP(p.Interface().(*bool), flag, shorthand, v.Bool(), desc)
			case reflect.Slice:
				switch f.Type.Elem().Kind() {
				case reflect.String:
					fset.StringSliceVarP(p.Interface().(*[]string), flag, shorthand, v.Interface().([]string), desc)
				}
			case reflect.Map:
				if f.Type.Key().Kind() == reflect.String {
					switch f.Type.Elem().Kind() {
					case reflect.String:
						fset.VarP(newStringMap(v.Interface().(map[string]string), p.Interface().(*map[string]string)), flag, shorthand, desc)
					case reflect.Int64:
						fset.VarP(newInt64Map(v.Interface().(map[string]int64), p.Interface().(*map[string]int64)), flag, shorthand, desc)
					}
				}
			}
		}

		pf := fset.Lookup(flag)
		if pf == nil {
			if f.Anonymous && v.Type().Kind() == reflect.Struct {
				if err := setup(fset, p.Interface(), "", ""); err != nil {
					return fmt.Errorf("failed to setup embedded struct: %w", err)
				}
			} else if v.Type().Kind() == reflect.Struct {
				if err := setup(fset, p.Interface(), env, flag); err != nil {
					return fmt.Errorf("failed to setup struct %T: %w", p.Interface(), err)
				}
			} else if v.Type().Kind() == reflect.Ptr && v.Elem().Type().Kind() == reflect.Struct {
				if err := setup(fset, p.Elem().Interface(), env, flag); err != nil {
					return fmt.Errorf("failed to setup struct pointer %T: %w", p.Elem().Interface(), err)
				}
			} else {
				return fmt.Errorf("unsupported type %s", f.Type.String())
			}
		} else {
			if err := viper.BindEnv(env); err != nil {
				return fmt.Errorf("failed to bind env %s: %w", env, err)
			}
			viper.SetDefault(env, v.Interface())
			if err := viper.BindPFlag(env, pf); err != nil {
				return fmt.Errorf("failed to bind env %s: %w", env, err)
			}
		}
	}

	return nil
}

func Parse(c interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           c,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			FlagValueHook(),
			StringMapHook(),
			Int64MapHook(),
		),
		TagName: "env",
	})

	if err != nil {
		return err
	}

	if err := decoder.Decode(viper.AllSettings()); err != nil {
		return err
	}

	return nil
}

func FlagValueHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		ptr := reflect.PtrTo(t)
		if !ptr.Implements(valueType) {
			return data, nil
		}

		v := reflect.New(t)
		pv := v.Interface().(pflag.Value)

		if err := pv.Set(data.(string)); err != nil {
			return data, nil
		}

		return v.Elem().Interface(), nil
	}
}
