package serial

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	translator "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func ValidateNil(object interface{}) bool {
	// 判断动态类型是否为nil [object=nil]
	if reflect.DeepEqual(object, nil) {
		return true
	}
	// 判断动态值是否为nil [object=(*int)(nil)]
	value := reflect.ValueOf(object)
	if value.Kind() == reflect.Ptr {
		return value.IsNil()
	}
	return false
}
func Validate(object interface{}, language string, translation bool) error {
	if ValidateNil(object) {
		return fmt.Errorf("validation object = %v are not allowed to be nil", object)
	}
	validate := validator.New()
	language = strings.TrimSpace(language)
	var translations translator.Translator
	if translation && len(language) > 0 {
		translations = universalTranslator(validate, language)
	}
	errs := validate.Struct(object)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			if translation && len(language) > 0 && !ValidateNil(translations) {
				return errors.New(err.Translate(translations))
			} else {
				return err
			}
		}
	}
	return nil
}
