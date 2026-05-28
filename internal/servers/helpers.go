package servers

import (
	"fmt"
	"net/http"
)

func (app *Application) backgroundTask(r *http.Request, fn func() error) {
	app.WG.Go(func() {
		defer func() {
			pv := recover()
			if pv != nil {
				app.Errors.ReportServerError(r, fmt.Errorf("%v", pv))
			}
		}()

		err := fn()
		if err != nil {
			app.Errors.ReportServerError(r, err)
		}
	})
}
