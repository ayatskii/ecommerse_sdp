package validator

import (
	"fmt"
	"regexp"
	"strings"
)

type CreditCardValidator struct{}

func NewCreditCardValidator() *CreditCardValidator {
	return &CreditCardValidator{}
}

func (v *CreditCardValidator) ValidateCardNumber(cardNumber string) error {

	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return fmt.Errorf("invalid card number length")
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(cardNumber) {
		return fmt.Errorf("card number must contain only digits")
	}

	sum := 0
	isEven := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	if sum%10 != 0 {
		return fmt.Errorf("invalid card number (failed Luhn check)")
	}

	return nil
}

func (v *CreditCardValidator) ValidateCVV(cvv string) error {
	if len(cvv) < 3 || len(cvv) > 4 {
		return fmt.Errorf("CVV must be 3 or 4 digits")
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(cvv) {
		return fmt.Errorf("CVV must contain only digits")
	}

	return nil
}

func (v *CreditCardValidator) ValidateExpiryDate(expiry string) error {
	if !regexp.MustCompile(`^\d{2}/\d{2}$`).MatchString(expiry) {
		return fmt.Errorf("expiry date must be in MM/YY format")
	}

	parts := strings.Split(expiry, "/")
	month := parts[0]
	year := parts[1]

	monthInt := 0
	fmt.Sscanf(month, "%d", &monthInt)

	if monthInt < 1 || monthInt > 12 {
		return fmt.Errorf("invalid month in expiry date")
	}

	yearInt := 0
	fmt.Sscanf(year, "%d", &yearInt)
	if yearInt < 0 {
		return fmt.Errorf("invalid year in expiry date")
	}

	return nil
}

type EmailValidator struct{}

func NewEmailValidator() *EmailValidator {
	return &EmailValidator{}
}

func (v *EmailValidator) Validate(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

type PhoneValidator struct{}

func NewPhoneValidator() *PhoneValidator {
	return &PhoneValidator{}
}

func (v *PhoneValidator) Validate(phone string) error {

	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "+", "")

	if len(phone) < 10 || len(phone) > 15 {
		return fmt.Errorf("invalid phone number length")
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(phone) {
		return fmt.Errorf("phone number must contain only digits")
	}

	return nil
}

type AmountValidator struct{}

func NewAmountValidator() *AmountValidator {
	return &AmountValidator{}
}

func (v *AmountValidator) Validate(amount float64, min, max float64) error {
	if amount < min {
		return fmt.Errorf("amount %.2f is below minimum %.2f", amount, min)
	}

	if amount > max {
		return fmt.Errorf("amount %.2f exceeds maximum %.2f", amount, max)
	}

	if amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	return nil
}

type CryptoAddressValidator struct{}

func NewCryptoAddressValidator() *CryptoAddressValidator {
	return &CryptoAddressValidator{}
}

func (v *CryptoAddressValidator) Validate(address, currency string) error {
	switch strings.ToUpper(currency) {
	case "BTC":
		return v.validateBitcoinAddress(address)
	case "ETH":
		return v.validateEthereumAddress(address)
	case "USDT":
		return v.validateEthereumAddress(address)
	default:
		return fmt.Errorf("unsupported cryptocurrency: %s", currency)
	}
}

func (v *CryptoAddressValidator) validateBitcoinAddress(address string) error {

	if len(address) < 26 || len(address) > 35 {
		return fmt.Errorf("invalid Bitcoin address length")
	}

	if !regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$|^bc1[a-z0-9]{39,59}$`).MatchString(address) {
		return fmt.Errorf("invalid Bitcoin address format")
	}

	return nil
}

func (v *CryptoAddressValidator) validateEthereumAddress(address string) error {

	if len(address) != 42 {
		return fmt.Errorf("invalid Ethereum address length")
	}

	if !regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`).MatchString(address) {
		return fmt.Errorf("invalid Ethereum address format")
	}

	return nil
}
