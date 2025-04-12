package helper

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validateInstance = validator.New()

func ValidatePayload(w http.ResponseWriter, model interface{}) error {
	err := validateInstance.Struct(model)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)

		var errorMessages []string
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed on the '%s' tag (value: '%v')", e.Field(), e.Tag(), e.Value()))
			}
		} else {
			errorMessages = append(errorMessages, err.Error())
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errorMessages,
		})
	}
	return err
}
