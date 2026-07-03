package utils

import (
	"strings"
	"testing"
)

type TestStruct struct {
	Name  string `validate:"required,min=3,max=10"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=18,lte=60"`
	Misc  string `validate:"alpha"`
}

func TestValidateStruct_Success(t *testing.T) {
	data := TestStruct{
		Name:  "Budi",
		Email: "budi@example.com",
		Age:   25,
		Misc:  "abc",
	}

	err := ValidateStruct(&data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidateStruct_Failure(t *testing.T) {
	data := TestStruct{
		Name:  "A",       // min error
		Email: "invalid", // email error
		Age:   10,        // gte error
		Misc:  "123",     // alpha error (default case)
	}

	err := ValidateStruct(&data)
	if err == nil {
		t.Errorf("Expected validation error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "minimal 3 karakter") {
		t.Errorf("Expected min error msg, got %s", errMsg)
	}
	if !strings.Contains(errMsg, "harus berupa email") {
		t.Errorf("Expected email error msg, got %s", errMsg)
	}
	if !strings.Contains(errMsg, "lebih besar atau sama dengan 18") {
		t.Errorf("Expected gte error msg, got %s", errMsg)
	}
	if !strings.Contains(errMsg, "tidak valid (alpha)") {
		t.Errorf("Expected default error msg, got %s", errMsg)
	}

	// Test required and max, lte
	data2 := TestStruct{
		Name:  "ThisIsTooLong", // max error
		Email: "",              // required error
		Age:   70,              // lte error
		Misc:  "abc",
	}
	err2 := ValidateStruct(&data2)
	if err2 == nil {
		t.Errorf("Expected validation error, got nil")
	}
	errMsg2 := err2.Error()
	if !strings.Contains(errMsg2, "maksimal 10 karakter") {
		t.Errorf("Expected max error msg, got %s", errMsg2)
	}
	if !strings.Contains(errMsg2, "tidak boleh kosong") {
		t.Errorf("Expected required error msg, got %s", errMsg2)
	}
	if !strings.Contains(errMsg2, "lebih kecil atau sama dengan 60") {
		t.Errorf("Expected lte error msg, got %s", errMsg2)
	}
}

func TestValidateStruct_NonValidationError(t *testing.T) {
	// struct without any validate tags but we pass an invalid type like nil to validate.Struct?
	// actually validator.Struct requires a struct, passing nil or non-struct returns an InvalidValidationError
	err := ValidateStruct(nil)
	if err == nil {
		t.Errorf("Expected error for nil, got nil")
	}
}
