// Package banner provides banner information
package banner

// Println print out banner information
func Println() {
	b := `
 =======================	
             _    ___ 
   _ __ _  _| |  / __|
  | '  \ || | |_| (_ |
  |_|_|_\_, |____\___|
        |__/          
 	
     My Looking Glass 
 ====== myLG v0.1.3 ======
	`
	println(b)
}
