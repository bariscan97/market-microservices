package utils

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)


func DynamicSearch(queryParams url.Values) (string, error) {
	
	query, gt, lt, flag := "", "0", "+inf", false
	
	for key, values := range queryParams {

		if key == "page" {
			continue
		}
		
		if key == "category"{
			query += fmt.Sprintf("@category:%s",  "{"  + strings.Join(values, "|") + "} ")
		}else if key == "lt" || key == "gt" {
			val := values[0]
			_, err := strconv.Atoi(val)
			if err != nil {
				log.Println("err: ", err)
				return "" , err
			}
			flag = true
			switch key {
			case "lt":
				lt = val
			case "gt":
				gt = val
			}
		}else {
			query += fmt.Sprintf("@%s:%s ", key, values[0])
		}
	
	}
	
	if flag {
		query += fmt.Sprintf("@%s:[%s %s] ", "price", gt, lt)
	}


	if query == "" {
		query = "*"
	}
	fmt.Println("QUERY: " , query)
	return query, nil
}