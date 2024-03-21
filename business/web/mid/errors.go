package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/liarsdice/business/web/auth"
	"github.com/ardanlabs/liarsdice/business/web/errs"
	"github.com/ardanlabs/liarsdice/foundation/logger"
	"github.com/ardanlabs/liarsdice/foundation/validate"
	"github.com/ardanlabs/liarsdice/foundation/web"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *logger.Logger) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := handler(ctx, w, r); err != nil {
				log.Error(ctx, "message", "msg", err)

				var er errs.Response
				var status int

				switch {
				case errs.IsTrusted(err):
					trsErr := errs.GetTrusted(err)

					if validate.IsFieldErrors(trsErr.Err) {
						fieldErrors := validate.GetFieldErrors(trsErr.Err)
						er = errs.Response{
							Error:  "data validation error",
							Fields: fieldErrors.Fields(),
						}
						status = trsErr.Status
						break
					}

					er = errs.Response{
						Error: trsErr.Error(),
					}
					status = trsErr.Status

				case auth.IsAuthError(err):
					er = errs.Response{
						Error: http.StatusText(http.StatusUnauthorized),
					}
					status = http.StatusUnauthorized

				default:
					er = errs.Response{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				if web.IsShutdown(err) {
					return err
				}
			}

			return nil
		}

		return h
	}

	return m
}
