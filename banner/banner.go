// Package banner provides banner information
package banner

import (
	"fmt"
)

const b = `
=================================================	
                          _    ___ 
                _ __ _  _| |  / __|
               | '  \ || | |_| (_ |
               |_|_|_\_, |____\___|
                      |__/          
	
                 My Looking Glass
           Free Network Diagnostic Tool
             www.facebook.com/mylg.io
                  http://mylg.io
================== myLG v%s ==================
	`

// Println print out banner information
func Println(version string) {
	fmt.Printf(b, version)
}
