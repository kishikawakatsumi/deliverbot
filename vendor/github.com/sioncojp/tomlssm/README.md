# tomlssm

tomlssm is extended TOML format with Amazon Simple System Manager.

## How to use

tomlssm has expand ssm value as macro prefixed by `"ssm://"`. See example below:

```
# config.toml

username = "test"
password = "ssm://password"
```

`"ssm://password"` should be set on you [System Manager Parameter Store](http://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html). If you set, tomlssm set a value stored in parameter store.

### Decode
```go
package main

import (
	"fmt"
	toml "github.com/sioncojp/tomlssm"
)

type Config struct {
	User     string `toml:"username"`
	Password string `toml:"password"`
}

func LoadToml(c string) (*Config, error) {
	var config Config
	if _, err := toml.Decode(c, &config, "ap-northeast-1"); err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	conf, err := LoadToml(`
username = "test"
password = "ssm://password"
`)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(conf)
}
```

### DecodeFile

```
# config.toml
username = "test"
password = "ssm://password"
```

```go
package main

import (
	"fmt"
	toml "github.com/sioncojp/tomlssm"
)

type Config struct {
	User     string `toml:"username"`
	Password string `toml:"password"`
}

func LoadToml(c string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(c, &config, "ap-northeast-1"); err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	conf, err := LoadToml("config.toml")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(conf)
}
```

## References

http://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html

# License

The MIT License

Copyright Shohei Koyama / sioncojp 

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
