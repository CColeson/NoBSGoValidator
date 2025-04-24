package validator

//
import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type HandlerFunc func(*ValidationContext)
type RuleFunc func(param []any) error

func ValidateStruct[T any](s T) error {
	typ := reflect.TypeOf(s)

	handler, ok := typeHandlers[reflect.TypeOf(s)]
	if !ok {
		panic("Type \"" + typ.Name() + "\" Hasn't been registered with RegisterType")
	}

	ctx := ValidationContext{}
	handler(s, &ctx)

	return ctx.err
}

var rules = make(map[string]RuleFunc, 0)
var typeHandlers = make(map[reflect.Type]func(a any, ctx *ValidationContext), 0)

func RegisterRule(ruleName string, fnc RuleFunc) {
	rules[ruleName] = fnc
}

func RegisterType[T any](handler func(s T, ctx *ValidationContext)) {
	typeHandlers[reflect.TypeFor[T]()] = func(a any, ctx *ValidationContext) {
		handler(a.(T), ctx)
	}
}

type ValidationContext struct {
	err error
}

func (ctx *ValidationContext) Message(message string) *ValidationContext {
	if ctx.err != nil {
		ctx.err = errors.New(message)
	}

	return ctx
}

func (ctx *ValidationContext) Check(handlerName string, params ...any) *ValidationContext {

	if ctx.err != nil {
		return ctx
	}

	rule, ok := rules[handlerName]
	if !ok {
		panic("Rule " + handlerName + " has not been registered")
	}

	err := rule(params)
	ctx.err = err
	return ctx
}

func (ctx *ValidationContext) MustErr(fnc func() error) *ValidationContext {
	if ctx.err != nil {
		return ctx
	}

	ctx.err = fnc()
	return ctx
}

func (ctx *ValidationContext) Must(fnc func() bool) *ValidationContext {
	if ctx.err != nil {
		return ctx
	}

	if !fnc() {
		ctx.err = errors.New("rule failed")
	}

	return ctx
}

func Init() {
	RegisterRule("notEmpty", func(param []any) error {
		for _, p := range param {
			str, ok := p.(string)
			if !ok {
				continue
			}

			if len(strings.TrimSpace(str)) == 0 {
				return errors.New("rule required failed")
			}
		}

		return nil
	})

	RegisterRule("greaterThan", func(params []any) error {
		// need at least two args: one comparer + at least one to compare
		if len(params) < 2 {
			msg := fmt.Sprintf("greaterThan: expected at least 2 parameters, got %d", len(params))
			panic(msg)
		}

		// --- determine the “comparer” from the first param ---
		first := params[0]
		rv := reflect.ValueOf(first)

		var comparer float64
		switch rv.Kind() {
		case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
			comparer = float64(rv.Len())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			comparer = float64(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			comparer = float64(rv.Uint())
		case reflect.Float32, reflect.Float64:
			comparer = rv.Float()
		default:
			return fmt.Errorf("greaterThan: unsupported type %T for comparer", first)
		}

		// --- for each of the remaining params, extract value/length and compare ---
		for i, arg := range params[1:] {
			rv := reflect.ValueOf(arg)

			var val float64
			switch rv.Kind() {
			case reflect.String, reflect.Array, reflect.Slice:
				val = float64(rv.Len())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				val = float64(rv.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				val = float64(rv.Uint())
			case reflect.Float32, reflect.Float64:
				val = rv.Float()
			default:
				return fmt.Errorf(
					"greaterThan: unsupported type %T at position %d",
					arg, i+2,
				)
			}

			if val <= comparer {
				return fmt.Errorf(
					"greaterThan: parameter at position %d (= %v) is not greater than %v",
					i+2, val, comparer,
				)
			}
		}

		return nil
	})

	RegisterRule("lessThan", func(params []any) error {
		// need at least two args: one comparer + at least one to compare
		if len(params) < 2 {
			msg := fmt.Sprintf("lessThan: expected at least 2 parameters, got %d", len(params))
			panic(msg)
		}

		// --- determine the “comparer” from the first param ---
		first := params[0]
		rv := reflect.ValueOf(first)

		var comparer float64
		switch rv.Kind() {
		case reflect.String, reflect.Array, reflect.Slice:
			comparer = float64(rv.Len())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			comparer = float64(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			comparer = float64(rv.Uint())
		case reflect.Float32, reflect.Float64:
			comparer = rv.Float()
		default:
			return fmt.Errorf("lessThan: unsupported type %T for comparer", first)
		}

		// --- for each of the remaining params, extract value/length and compare ---
		for i, arg := range params[1:] {
			rv := reflect.ValueOf(arg)

			var val float64
			switch rv.Kind() {
			case reflect.String, reflect.Array, reflect.Slice:
				val = float64(rv.Len())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				val = float64(rv.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				val = float64(rv.Uint())
			case reflect.Float32, reflect.Float64:
				val = rv.Float()
			default:
				return fmt.Errorf(
					"lessThan: unsupported type %T at position %d",
					arg, i+2,
				)
			}

			if val >= comparer {
				return fmt.Errorf(
					"lessThan: parameter at position %d (= %v) is not less than %v",
					i+2, val, comparer,
				)
			}
		}

		return nil
	})

}
