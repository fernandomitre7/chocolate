package reqbody

import (
	"io/ioutil"
	"net/http"

	"chocolate/service/api/shared/apierror"
	"chocolate/service/models"
)

// Read reads the body of an http request into a provided APIObject
func Read(r *http.Request, out models.APIObject) (apierr *apierror.Error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apierr = apierror.New(http.StatusInternalServerError, err.Error(), apierror.CodeInternalReadBody)
		return
	}
	if err = out.Decode(body); err != nil {
		apierr = apierror.New(http.StatusBadRequest, err.Error(), apierror.CodeBadRequestBody)
	}
	return
}
