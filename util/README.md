# util
Utility package

List of utility
- Value lookup 
- Value matcher
- UID Generator
- ImgProxy URL Generator

## Usage

### Value lookup & matcher

```go
package main

import (
    "fmt"

	"github.com/diki-haryadi/govega/util"
)

func main() {
	obj := map[string]interface{}{
		"result": map[string]interface{}{
			"name": "SiCepat",
		},
	}

    if util.FieldExist("result.name", obj) {
        fmt.Println("OK")
    }

	name, ok := util.Lookup("result.name", obj)
    if name == "SiCepat" {
        fmt.Println("OK")
    }

    if util.Match("result.name", obj, "SiCepat") {
        fmt.Println("OK")
    }

}
```

### UID Generator

```go
package main

import (
    "fmt"

	"github.com/diki-haryadi/govega/util"
)

func main() {
    var id int64
	id = GenerateRandUID()
    id = GenerateSeqUID()

    uid := NewUIDRandomNum()
    id = uid.Generate()

    uid = NewUIDSequenceNum()
    id = uid.Generate()

    fmt.Println(EncodeUID(id))

}
```

### Snowflake Generator

```go
package main

import (
    "fmt"
    "github.com/diki-haryadi/govega/util"
)

func main() {
	pod, err := NewPOD(&SnowflakeOpts{
		Epoch: 1626670674000, //Set in config
		POD:   1, //Set in config
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

### ImgProxy URL Generator

```go
package main

import (
    "fmt"

	"github.com/diki-haryadi/govega/util"
)

func main() {
	img := NewImageProxy("http://localhost:8080", "https://sicepatresi.s3.amazonaws.com/0016100/001610004980.jpg", "testkey12345", "testkey12345")
	url, err := img.GetURL()
	if err != nil {
		panic(err) 
	}
	fmt.Println(url) //http://localhost:8080/oOXB13_3e8ZR1PTQQLcpI6mHACY7qgvf7GjzypnQOqs/rs:fit:1024:768:0/g:no/aHR0cHM6Ly9zaWNlcGF0cmVzaS5zMy5hbWF6b25hd3MuY29tLzAwMTYxMDAvMDAxNjEwMDA0OTgwLmpwZw.jpg
}
```