# Gohm ॐ ![build](https://travis-ci.org/pote/gohm.svg)

Gohm is a Go port of the popular [Ohm](https://github.com/soveran/ohm) Ruby library, it provides a simple interface to store and retrieve your model data in a Redis database.

Gohm implements nothing but the basic usage right now, but expect all or most features in Ohm to be implemented into Gohm as time goes by, contributions are very welcome. :) 

## The Basics


```go
package main

import(
  "github.com/pote/gohm"
)

type User struct{
	ID    string `ohm:"id"`
	Name  string `ohm:"name"`
	Email string `ohm:"email"`
}

func main() {
 	Gohm, _ := gohm.NewGohm()

  	u := User{
  		Name: "Marty",
		Email: "marty@mcfly.com",
  	}

  	Gohm.Save(&u)

  	u.ID //=> "1"

  	u2 := User{ID: "1"}
  	Gohm.Load(&u2)

  	u2.Name //=> "Marty"
}
```

## Ohm compatibility

Both Ohm and Gohm are powered by [ohm-scripts](https://github.com/soveran/ohm-scripts), a set of Lua scripts that bundle common operations and make it easy to write a port such as this one, it also means that by adhering to the ohm standard **models stored with Gohm can be loaded from Ohm, and vice-versa**.
