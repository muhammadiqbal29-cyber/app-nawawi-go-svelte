package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validate adalah singleton validator instance
var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

// ValidateStruct memvalidasi struct dan memformat error menjadi teks tunggal yang mudah dibaca.
func ValidateStruct(data interface{}) error {
	err := Validate.Struct(data)
	if err == nil {
		return nil
	}

	var errorMessages []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			// Customize pesan error berdasarkan tag
			field := e.Field()
			switch e.Tag() {
			case "required":
				errorMessages = append(errorMessages, fmt.Sprintf("Kolom %s tidak boleh kosong", field))
			case "email":
				errorMessages = append(errorMessages, fmt.Sprintf("Kolom %s harus berupa email yang valid", field))
			case "min":
				errorMessages = append(errorMessages, fmt.Sprintf("Kolom %s minimal %s karakter", field, e.Param()))
			case "max":
				errorMessages = append(errorMessages, fmt.Sprintf("Kolom %s maksimal %s karakter", field, e.Param()))
			case "gte":
				errorMessages = append(errorMessages, fmt.Sprintf("Kolom %s harus lebih besar atau sama dengan %s", field, e.Param()))
			case "lte":
				errorMessages = append(errorMessages, fmt.Sprintf("Kolom %s harus lebih kecil atau sama dengan %s", field, e.Param()))
			default:
				errorMessages = append(errorMessages, fmt.Sprintf("Kolom %s tidak valid (%s)", field, e.Tag()))
			}
		}
	} else {
		// Error bukan dari validator
		return err
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("%s", strings.Join(errorMessages, ", "))
	}
	return nil
}
