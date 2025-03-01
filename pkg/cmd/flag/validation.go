package flag

import (
	flagutil "github.com/redhat-developer/app-services-cli/pkg/cmdutil/flags"
)

// ValidateOutput checks if value v is a valid value for --output
func ValidateOutput(v string) error {
	isValid := flagutil.IsValidInput(v, flagutil.ValidOutputFormats...)

	if isValid {
		return nil
	}

	return InvalidValueError("output", v, flagutil.ValidOutputFormats...)
}
