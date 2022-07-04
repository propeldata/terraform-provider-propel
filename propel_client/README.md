# CMS GraphQL Client

### Client generation

Download the client generation tool:

```shell
go get github.com/Khan/genqlient
```

In this folder:

```shell
go run github.com/Khan/genqlient
```

### Usage

```go
package my_package

import (
	"context"
	"fmt"
	"log"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func main() {
	c, err := pc.NewPropelClient("clientID", "clientSecret", "user-agent")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := pc.DataSource(context.Background(), c, "DSO00000000000000000000000000")
	
	fmt.Println(resp.DataSource.Account.Id, err)
}
```
