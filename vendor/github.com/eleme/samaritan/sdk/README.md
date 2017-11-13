# Sam SDK Go

Samaritan SDK for Go.

## Example

```Go
// I have an application, its appID is "arch.sash", and it runs in cluster "channel-stable-1".
// I need use the service provided by "arch.q" and its cluster is "overall.xg".

package main
import "github.com/eleme/samaritan/sdk"

func main() {
    // Create client with my app's AppID and cluster.
    c, err := sdk.NewLocalClient("arch.sash", "channel-stable-1")
    if err != nil {
        // Handle err.
    }
    // Get host and port of arch.q's overall.xg.
    host, port, err := c.GetHostPort("arch.q", "overall.xg")
}
```
