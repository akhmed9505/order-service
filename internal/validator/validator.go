package validator

import (
	"regexp"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// initValidator initializes the validator and registers custom validations
func initValidator() {
	validate = validator.New()

	// UUID validation: must match standard UUID format
	_ = validate.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		re := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
		return re.MatchString(value)
	})

	// E.164 phone number validation: must start with + and contain 10-15 digits
	_ = validate.RegisterValidation("e164", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		re := regexp.MustCompile(`^\+[1-9]\d{9,15}$`)
		return re.MatchString(phone)
	})

	// ISO4217 currency code validation: must be 3 uppercase letters
	_ = validate.RegisterValidation("iso4217", func(fl validator.FieldLevel) bool {
		cur := fl.Field().String()
		re := regexp.MustCompile(`^[A-Z]{3}$`)
		return re.MatchString(cur)
	})
}

// ValidateOrder validates the entire order struct including nested fields and slices
func ValidateOrder(order interface{}) error {
	once.Do(initValidator)
	return validate.Struct(order)
}
