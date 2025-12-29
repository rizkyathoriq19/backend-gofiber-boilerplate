package validator

import (
	"testing"
)

func TestValidateStruct(t *testing.T) {
	type TestUser struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
		Name     string `json:"name" validate:"required,min=2,max=100"`
	}

	tests := []struct {
		name    string
		input   TestUser
		wantErr bool
	}{
		{
			name: "Valid input",
			input: TestUser{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "John Doe",
			},
			wantErr: false,
		},
		{
			name: "Invalid email",
			input: TestUser{
				Email:    "invalid-email",
				Password: "password123",
				Name:     "John Doe",
			},
			wantErr: true,
		},
		{
			name: "Missing email",
			input: TestUser{
				Email:    "",
				Password: "password123",
				Name:     "John Doe",
			},
			wantErr: true,
		},
		{
			name: "Password too short",
			input: TestUser{
				Email:    "test@example.com",
				Password: "123",
				Name:     "John Doe",
			},
			wantErr: true,
		},
		{
			name: "Name too short",
			input: TestUser{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "J",
			},
			wantErr: true,
		},
		{
			name: "All fields empty",
			input: TestUser{
				Email:    "",
				Password: "",
				Name:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatValidationError(t *testing.T) {
	type TestInput struct {
		Email string `json:"email" validate:"required,email"`
	}

	// Test with invalid email
	input := TestInput{Email: "invalid"}
	err := ValidateStruct(input)

	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	formatted := FormatValidationErrorForResponseBilingual(err)

	if len(formatted) == 0 {
		t.Error("Expected formatted errors, got empty slice")
	}

	// Check that the error contains field info
	foundEmailError := false
	for _, e := range formatted {
		if e.Field == "email" {
			foundEmailError = true
			if e.Message.EN == "" {
				t.Error("Expected English message, got empty")
			}
			if e.Message.ID == "" {
				t.Error("Expected Indonesian message, got empty")
			}
		}
	}

	if !foundEmailError {
		t.Error("Expected error for 'email' field")
	}
}

func TestValidateOptionalFields(t *testing.T) {
	type OptionalUser struct {
		Name  string `json:"name" validate:"omitempty,min=2"`
		Phone string `json:"phone" validate:"omitempty,min=10"`
	}

	tests := []struct {
		name    string
		input   OptionalUser
		wantErr bool
	}{
		{
			name: "All empty (valid - optional)",
			input: OptionalUser{
				Name:  "",
				Phone: "",
			},
			wantErr: false,
		},
		{
			name: "Valid name, empty phone",
			input: OptionalUser{
				Name:  "John",
				Phone: "",
			},
			wantErr: false,
		},
		{
			name: "Name too short",
			input: OptionalUser{
				Name:  "J",
				Phone: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	type UUIDInput struct {
		ID string `json:"id" validate:"required,uuid"`
	}

	tests := []struct {
		name    string
		input   UUIDInput
		wantErr bool
	}{
		{
			name:    "Valid UUID",
			input:   UUIDInput{ID: "550e8400-e29b-41d4-a716-446655440000"},
			wantErr: false,
		},
		{
			name:    "Invalid UUID",
			input:   UUIDInput{ID: "not-a-uuid"},
			wantErr: true,
		},
		{
			name:    "Empty UUID",
			input:   UUIDInput{ID: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
