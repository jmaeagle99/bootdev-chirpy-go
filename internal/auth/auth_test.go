package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	defaultUserId := uuid.MustParse("b92b8014-6b95-42df-99d6-1e78551bb6ea")
	defaultTokenSecret := "SGVsbG8sIFdvcmxkIQ=="
	defaultExpiry, _ := time.ParseDuration("5s")
	jwt1, _ := MakeJWT(defaultUserId, defaultTokenSecret, defaultExpiry)
	jwt2, _ := MakeJWT(defaultUserId, "VW5leHBlY3RlZCB0b2tlbiBzZWNyZXQ=", defaultExpiry)

	tests := []struct {
		name           string
		tokenString    string
		decodeSecret   string
		expectedError  bool
		expectedUserId uuid.UUID
	}{
		{
			name:           "Valid token with expected subject",
			tokenString:    jwt1,
			decodeSecret:   defaultTokenSecret,
			expectedError:  false,
			expectedUserId: defaultUserId,
		},
		{
			name:           "Mismatched signature",
			tokenString:    jwt2,
			decodeSecret:   defaultTokenSecret,
			expectedError:  true,
			expectedUserId: uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualUserId, err := ValidateJWT(tt.tokenString, tt.decodeSecret)
			if (err != nil) != tt.expectedError {
				t.Errorf("ValidateJWT() error = %v, expectedError %v", err, tt.expectedError)
			}
			if actualUserId != tt.expectedUserId {
				t.Errorf("ValidateJWT() expects %v, got %v", tt.expectedUserId, actualUserId)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		headers       http.Header
		expectedError bool
		expectedValue string
	}{
		{
			name:          "No headers",
			headers:       http.Header{},
			expectedError: true,
			expectedValue: "",
		},
		{
			name:          "Missing Authorization header",
			headers:       http.Header{"Content-Type": []string{"application/json"}},
			expectedError: true,
			expectedValue: "",
		},
		{
			name:          "Empty Authorization header",
			headers:       http.Header{"Authorization": []string{}},
			expectedError: true,
			expectedValue: "",
		},
		{
			name:          "Authorization header mismatch token type",
			headers:       http.Header{"Authorization": []string{"Custom ABCDE"}},
			expectedError: true,
			expectedValue: "",
		},
		{
			name:          "Authorization header is Bearer",
			headers:       http.Header{"Authorization": []string{"Bearer XYZ"}},
			expectedError: false,
			expectedValue: "XYZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualValue, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.expectedError {
				t.Errorf("GetBearerToken() error = %v, expectedError %v", err, tt.expectedError)
			}
			if actualValue != tt.expectedValue {
				t.Errorf("GetBearerToken() expects %v, got %v", tt.expectedValue, actualValue)
			}
		})
	}
}
