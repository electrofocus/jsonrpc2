package rpc

import "fmt"

func Subject(client, subject string) string {
	return fmt.Sprintf("%s.%s", client, subject)
}
