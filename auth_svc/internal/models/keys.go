package models

import "fmt"

func AuthAPIserverKey(clientID string) string {
	return fmt.Sprintf("apiserver#%s", clientID)
}

func AuthUserKey(userID string) string {
	return fmt.Sprintf("userid#%s", userID)
}
