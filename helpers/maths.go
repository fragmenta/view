package helpers

import (
	"fmt"
	"strings"
    "strconv"
)




// PRICES

// Convenience function returning price in pence as string for use in params
func PriceToCentsString(p string) string {
    if p == "" {
        return "0"// Return 0 for blank price
    } else {
        return fmt.Sprintf("%d",PriceToCents(p))
    }
}

// Convert a price in human friendly notation (£45 or £34.40) to a price in pence as an int64
func PriceToCents(p string) int {
    price := strings.Replace(p,"£","",-1)
    price = strings.Replace(price,",","",-1)// assumed to be in thousands
    price = strings.Replace(price," ","",-1)
        
    var pennies int
    var err error
    if strings.Contains(price,".") {
       // Split the string on . and rejoin with padded pennies
       parts := strings.Split(price,".")
       
       if len(parts[1]) == 0 {
           parts[1] = "00"
        }else if len(parts[1]) == 1 {
           parts[1] = parts[1] + "0"
       }  
       
       price = parts[0] + parts[1]
       
       pennies,err = strconv.Atoi(price)
    } else {
       pennies,err = strconv.Atoi(price)
       pennies = pennies * 100  
    }
    if err != nil {
        fmt.Printf("Error converting price %s",price)
        pennies = 0;
    }

    return pennies
}


// Convert a price in pence to a human friendly price - NB We DO include currency in £
// we'll need to adjust this to take a currency setting later
func CentsToPrice(p int64) string {
    price := fmt.Sprintf("£%.2f",float64(p)/100.0)
    return strings.TrimSuffix(price,".00")// remove zero pence at end if we have it
}



func Mod(a int,b int) int {
    return a % b
}

func Add(a int,b int) int {
    return a + b
}

func Odd(a int) bool {
    return a % 2 == 0
}