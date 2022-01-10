package Exception

import "log"

func ErrorHandler(exception error){
	if exception != nil{
		log.Fatal(exception)
	}
}
