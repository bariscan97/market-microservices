package utils

import (
	"fmt"
)

func SqlUpdateQuery(fields map[string]interface{}, table string, conditions map[string]interface{}) (string, []interface{}) {

	sql := fmt.Sprintf("UPDATE %s SET", table)

	condition := " WHERE "

	size := len(fields)

	count := 0

	parameters := make([]interface{}, 0)

	for key, value := range fields {
		s := fmt.Sprintf(" %s = $%v ", key, count+1)
		parameters = append(parameters, value)
		sql += s
		count++
		if count < size {
			sql += ","
		}
	}

	for key, value := range conditions {
		s := fmt.Sprintf(" %s = $%v ", key, count+1)
		parameters = append(parameters, value)
		condition += s
		count++
		if count < size+len(conditions) {
			condition += " AND "
		}
	}
	sql += condition

	return sql, parameters
}
