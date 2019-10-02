package extras

type Prefix string

const (
	SessionPrefix        Prefix = "sess:"
	ConfirmAccountPrefix Prefix = "conf:"
	ResetPasswordPrefix  Prefix = "pw:"
)

type EmailTemplates string

const (
	ConfirmAccount EmailTemplates = "confirmaccount.html"
	ResetPassword  EmailTemplates = "resetpassword.html"
)

type ErrorMsg string

const (
	BadTokenError       ErrorMsg = "Token does not exist or has expired"
	WrongUserOrPassword ErrorMsg = "Username and/or password is incorrect"
	AccessDenied        ErrorMsg = "Access denied. Please login to get access to this data"
)
