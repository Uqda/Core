package main

import (
	"fmt"
	"strings"

	"github.com/Uqda/Core/src/admin"
)

func validateAdminArguments(args []string, available admin.ListResponse) error {
	for _, entry := range available.List {
		if !strings.EqualFold(entry.Command, args[0]) {
			continue
		}
		for _, arg := range args[1:] {
			key, _, _ := strings.Cut(arg, "=")
			valid := false
			for _, field := range entry.Fields {
				if key == field {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("%s does not accept %q; run 'uqda commands' for available arguments", args[0], key)
			}
		}
		return nil
	}
	return fmt.Errorf("command %q is not available on this node; run 'uqda commands'", args[0])
}
