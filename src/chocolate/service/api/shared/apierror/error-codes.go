package apierror

// Code are the codes for more specific information about errors
type Code string

const (
	// CodeUnknown is the default code
	CodeUnknown = Code("0000")
	// CodeInternal = Internal Server Error Generic error
	CodeInternal = Code("0001")
	// CodeInternalJWT = Internal Error Generating a JWT
	CodeInternalJWT = Code("0002")
	// CodeInternalReadBody = Internal Error, could't read request body
	CodeInternalReadBody = Code("0003")
	// CodeInternalEmail = Internal Error that prevented system from sending or using email service
	CodeInternalEmail = Code("0004")
	// CodeInternalDB = Internal Error related DB
	CodeInternalDB = Code("0010")
	// CodeUnauth = Unauthorized
	CodeUnauth = Code("0100")
	// CodeUnauthMalformed = Unauthorized because JWT was malformed
	CodeUnauthMalformed = Code("0102")
	// CodeUnauthExpired = Unauthorized because JWT expired
	CodeUnauthExpired = Code("0103")
	// CodeUnauthNotActive = Unauthorized because JWT is not active yet
	CodeUnauthNotActive = Code("104")
	// CodeForbidden = Forbidden
	CodeForbidden = Code("0110")
	// CodeForbiddenNotConfirmed = User is OK but email is not confirmed
	CodeForbiddenNotConfirmed = Code("0111")
	// CodeBadRequest = Bad Request generic error code
	CodeBadRequest            = Code("0200")
	CodeBadReqPasswordConfirm = Code("0201")
	// CodeBadRequestBody = Bad Request because body was wrong
	CodeBadRequestBody = Code("0204")
	// CodeBadRequestParams = Ban Request Query params
	CodeBadRequestParams = Code("0205")
	// CodeResourceNotFound = Resource doesnt exists
	CodeResourceNotFound = Code("0301")
)
