package ui

import (
	"fmt"
	"net/http"
)

func (c *Controller) isStructOperationAllowed(r *http.Request, structName string, op int) bool {
	allowedTypes := r.Context().Value(ContextValue(fmt.Sprintf("AllowedTypes_%d", op)))
	if allowedTypes != nil {
		v, ok := allowedTypes.(map[string]bool)[structName]
		if !ok || !v {
			v2, ok2 := allowedTypes.(map[string]bool)["all"]
			if !ok2 || !v2 {
				return false
			}
		}
	}

	return true
}
