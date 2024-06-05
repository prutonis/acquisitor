/*
Copyright © 2024 IonM
*/
package srv

import (
	"fmt"
	"log"
)

func Init() {
	fmt.Println("srv init called")
}

func Serve() error {
	if 1 != 1 {
		return fmt.Errorf("serve error")
	}
	return nil
}

func StartServer() {
	Init()
	log.Fatal("Couldn't start server", Serve())

}
