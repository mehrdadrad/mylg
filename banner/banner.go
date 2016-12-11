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
                  http://mylg.io
================== myLG v%s ==================
	`

// Println print out banner information
func Println(version string) {
	fmt.Printf("\033[2J")           // clear screen
	fmt.Printf("\033[%d;%dH", 0, 0) // move cursor to x-0, y=0
	fmt.Printf(b, version)          // print banner including version
}
