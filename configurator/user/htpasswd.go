package user

import (
	"os"

	"github.com/foomo/htpasswd"
)

// ReadHTTPUsers func read list of users from .htpasswd file
func ReadHTTPUsers() []PMMUser {
	var users []PMMUser
	if userMap, err := htpasswd.ParseHtpasswdFile(SSMConfig.HtpasswdPath); err == nil {
		for username := range userMap {
			users = append(users, PMMUser{Username: username, Password: "********"})
		}
	}
	return users
}

func createHTTPUser(newUser PMMUser) error {
	// htpasswd.HashBCrypt is better, but nginx server in CentOS 7, doesn't support it :(
	return htpasswd.SetPassword(SSMConfig.HtpasswdPath, newUser.Username, newUser.Password, htpasswd.HashSHA)
}

func deleteHTTPUser(username string) error {
	if err := htpasswd.RemoveUser(SSMConfig.HtpasswdPath, username); err != nil {
		return err
	}

	fi, err := os.Stat(SSMConfig.HtpasswdPath)
	if err != nil {
		return err
	}

	if fi.Size() == 0 {
		return os.Remove(SSMConfig.HtpasswdPath)
	}
	return nil
}
