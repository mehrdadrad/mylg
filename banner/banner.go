// Package banner provides banner information
package banner

import (
	"fmt"
)

// Println print out banner information
func Println() {
	b := `
=================================================	
                          _    ___ 
                _ __ _  _| |  / __|
               | '  \ || | |_| (_ |
               |_|_|_\_, |____\___|
                      |__/          
	
                 My Looking Glass
                  http://mylg.io
================== myLG v0.1.7 ==================
	`
	fmt.Println(b)
}
