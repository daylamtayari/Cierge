package main

import (
	"errors"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var userPasswordCmd = &cobra.Command{
	Use:   "password",
	Short: "Change user password",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()

		var oldPassword string
		var newPassword string
		err := runHuh(huh.NewInput().Title("Enter current password:").EchoMode(huh.EchoModePassword).Value(&oldPassword))
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to prompt user for current password")
		}
		err = runHuh(huh.NewInput().Title("Enter new password:").EchoMode(huh.EchoModePassword).Value(&newPassword).Validate(validateNewPassword))
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to prompt user for new password")
		}

		err = client.ChangePassword(oldPassword, newPassword)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to change password")
		}
		logger.Info().Msg("Successfully changed password")
	},
}

func validateNewPassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < 8 {
		return errors.New("Password must be at least 8 characters") //nolint:staticcheck
	}
	if length > 128 {
		return errors.New("Password must be at most 128 characters") //nolint:staticcheck
	}

	var hasLetter, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasLetter {
		return errors.New("Password must contain at least one letter") //nolint:staticcheck
	}
	if !hasDigit {
		return errors.New("Password must contain at least one digit") //nolint:staticcheck
	}
	if !hasSpecial {
		return errors.New("Password must contain at least one special character") //nolint:staticcheck
	}
	return nil
}
