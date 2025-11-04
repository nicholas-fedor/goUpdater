package http

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// ValidateResponse validates the HTTP response by checking for nil response
// and ensuring the status code indicates success (2xx range).
// It returns an error if the response is invalid or unsuccessful.
func ValidateResponse(resp *http.Response) error {
	if resp == nil {
		return ErrResponseNil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d: %w", resp.StatusCode, ErrUnexpectedStatus)
	}

	return nil
}

// CheckStatusCode checks if the response status code matches the expected code.
// It returns an error if the codes do not match.
func CheckStatusCode(resp *http.Response, expected int) error {
	if resp.StatusCode != expected {
		return fmt.Errorf("%w: expected %d, got %d", ErrStatusCodeMismatch, expected, resp.StatusCode)
	}

	return nil
}

// ReadBody reads the entire response body and returns it as a byte slice.
// It ensures the response body is closed after reading.
// Returns an error if reading fails.
func ReadBody(resp *http.Response) ([]byte, error) {
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// ParseJSON parses the response body as JSON into the provided value.
// It reads the body, closes it, and unmarshals the JSON.
// The value must be a pointer to the target struct.
// Returns an error if reading or unmarshaling fails.
func ParseJSON(resp *http.Response, value interface{}) error {
	body, err := ReadBody(resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, value)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}

// ParseXML parses the response body as XML into the provided value.
// It reads the body, closes it, and unmarshals the XML.
// The value must be a pointer to the target struct.
// Returns an error if reading or unmarshaling fails.
func ParseXML(resp *http.Response, value interface{}) error {
	body, err := ReadBody(resp)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(body, value)
	if err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	return nil
}
