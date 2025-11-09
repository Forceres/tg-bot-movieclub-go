package telegraph

import (
	"fmt"
	"net/http"
	"time"

	"github.com/celestix/telegraph-go/v2"
)

type Telegraph struct {
	Client *telegraph.TelegraphClient
	Account *telegraph.Account
}

func InitTelegraph() (*Telegraph, error) {
	client := telegraph.GetTelegraphClient(&telegraph.ClientOpt{
		HttpClient: &http.Client{
			Timeout: 6 * time.Second,
		},
	})

	account, err := client.CreateAccount("telegraph-go", &telegraph.CreateAccountOpts{
		AuthorName: "KinoClassBot",
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &Telegraph{Client: client, Account: account}, nil
}