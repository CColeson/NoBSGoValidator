package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type ValidationContext struct {
	validator *Validator
	err       error
}
type HandlerFunc func(a any, ctx *ValidationContext)
type RuleFunc func(param []any) error
type Validator struct {
	rules        map[string]RuleFunc
	typeHandlers map[reflect.Type]HandlerFunc
}

func RegisterRule(v *Validator, ruleName string, fnc RuleFunc) {
	v.rules[ruleName] = fnc
}

func RegisterType[T any](v *Validator, handler func(s T, ctx *ValidationContext)) {
	v.typeHandlers[reflect.TypeFor[T]()] = func(a any, cc *ValidationContext) {
		handler(a.(T), cc)
	}
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

	rule, ok := ctx.validator.rules[handlerName]
	if !ok {
		panic("Rule " + handlerName + " has not been registered to specified validator")
	}

	err := rule(params)
	ctx.err = err
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

func New() *Validator {
	validator := &Validator{}
	RegisterRule(validator, "notEmpty", func(param []any) error {
		for _, p := range param {

			length := 0
			rv := reflect.ValueOf(p)
			kind := rv.Kind()

			if kind == reflect.Array ||
				kind == reflect.Slice ||
				kind == reflect.Map {
				length = rv.Len()
			} else if kind == reflect.String {
				length = len(strings.TrimSpace(p.(string)))
			}

			if length == 0 {
				return errors.New("required rule failed")
			}
		}

		return nil
	})

	RegisterRule(validator, "greaterThan", func(params []any) error {
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

	RegisterRule(validator, "lessThan", func(params []any) error {
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

	RegisterRule(validator, "isEmail", func(param []any) error {
		if len(param) == 0 {
			panic("No parameters passed to email rule")
		}

		email, ok := param[0].(string)
		if !ok {
			panic("Parameter passed to email rule is not a string")
		}

		var ampIsThere bool
		var spacesThere bool
		var textBeforeAmp bool
		var textAfterAmp bool
		var dotAfterAmp bool
		var otherError bool

		for _, r := range email {
			if r == '@' {
				if ampIsThere {
					otherError = true
				}

				ampIsThere = true
			} else if !ampIsThere {
				textBeforeAmp = true
			} else if r == '.' {
				dotAfterAmp = true
			} else {
				textAfterAmp = true
			}

			if r == ' ' || r == ',' {
				spacesThere = true
			}
		}

		if spacesThere ||
			!ampIsThere ||
			!textAfterAmp ||
			!textBeforeAmp ||
			!dotAfterAmp ||
			otherError {
			return errors.New("Email addresses must be valid, working, and must have no commas or spaces")
		}

		return nil
	})

	return validator
}
