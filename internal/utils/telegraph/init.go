package telegraph

import (
	"fmt"
	"net/http"
	"time"

	"github.com/celestix/telegraph-go/v2"
)

func InitTelegraph() (*telegraph.TelegraphClient, error) {
	client := telegraph.GetTelegraphClient(&telegraph.ClientOpt{
		HttpClient: &http.Client{
			Timeout: 6 * time.Second,
		},
	})

	// Use this method to create account
	_, err := client.CreateAccount("telegraph-go", &telegraph.CreateAccountOpts{
		AuthorName: "KinoClassBot",
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return client, nil
}