////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"code.google.com/p/go.crypto/bcrypt"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: hashpassword <password>\n")
		fmt.Printf("Then on the server, in psql, run something like:")
		fmt.Printf("update \"users\" set passwordhash='\\x2432612431302433666e424a5a794d622e623450334539655832585765333164336b6a5855684c49474d76354543456a365a39386d6d6d6132726c65' where id=1;")
		return
	}
	password := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Couldn't hash password: %v", err)
		return
	}

	fmt.Printf("%s\n", hex.EncodeToString(hash))
}
