### Snowflake Generator

```go
package main

import (
    "fmt"
)

func main() {
	pod, err := NewPOD(&SnowflakeOpts{
		Epoch: 1626670674000, //Set in configs
		POD:   1, //Set in configs
	})
	if err != nil {
		//Handler error should be fatal or else
	}

	//Generate POD id
	id := pod.Generate()
	
	fmt.Printf("Int64    : %#v", id.Int64()) //31677837479936
	fmt.Printf("String   : %#v", id.String()) //"31677837479936"
	fmt.Printf("Base2    : %#v", id.Base2()) //"111001100111110010010010000000001000000000000"
	fmt.Printf("Base32   : %#v", id.Base32()) //"h36jryryy"
	fmt.Printf("Base36   : %#v", id.Base36()) //"b88lijda8"
	fmt.Printf("Base58   : %#v", id.Base58()) //"fm892iv3"
	fmt.Printf("Base64   : %#v", id.Base64()) //"MzE2Nzc4Mzc0Nzk5MzY="
	fmt.Printf("Bytes    : %#v", id.Bytes()) //[]byte{0x33, 0x31, 0x36, 0x37, 0x37, 0x38, 0x33, 0x37, 0x34, 0x37, 0x39, 0x39, 0x33, 0x36}
	fmt.Printf("IntBytes : %#v", id.IntBytes()) //[8]uint8{0x0, 0x0, 0x1c, 0xcf, 0x92, 0x40, 0x10, 0x0}
}
```