package extras

// Prefix is a prefix type for prefixing things
type Prefix string

const (
	// SessionPrefix for prefixing the redisdatabase
	RefreshTokenPrefix Prefix = "rt:"
	AccessTokenPrefix  Prefix = "access:"
	// ConfirmAccountPrefix for prefixing the redisdatabase
	ConfirmAccountPrefix Prefix = "confirm:"
	// ResetPasswordPrefix for prefixing the redisdatabase
	ResetPasswordPrefix Prefix = "reset:"
)

// EmailTemplates holds all the filenames for templates used for sending out emails to users
type EmailTemplates string

const (
	// ConfirmAccount holds the filename for the template used for sending out confirmation emails
	ConfirmAccount EmailTemplates = "confirmaccount.html"
)

// ErrorMsg is self explanatory
type ErrorMsg string

const (
	// BadTokenError will be shown when a token does not exist or is expired
	BadTokenError ErrorMsg = "Token does not exist or has expired"
	// WrongUserOrPassword is the standardized message shown when a login is unsuccessful
	WrongUserOrPassword ErrorMsg = "Username and/or password is incorrect"
	// AccessDenied will be shown when the user doesn't have sufficient access to the data he/she wants
	AccessDenied ErrorMsg = "Access denied. Please login to get access to this data"
)
