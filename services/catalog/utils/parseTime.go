package utils

import (
	"time"
	"fmt"
    "strconv"
)


func ConvertToUnix(value interface{}) (int64, error) {
	str, ok := value.(string)
	if !ok {
		return 0, fmt.Errorf("veri string formatında değil")
	}

	parsedTime, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return 0, err
	}

	return parsedTime.Unix(), nil
}

func ConvertToTime(value interface{}) (time.Time, error) {
    strValue, ok := value.(string)
    if !ok {
        return time.Time{}, fmt.Errorf("veri string formatinda değil")
    }
    
    parsedTime, err := time.Parse(time.RFC3339, strValue)
    if err != nil {
        return time.Time{}, fmt.Errorf("tarih dönüştürme hatasi: %v", err)
    }
    
    return parsedTime, nil
}

func ConvertUnixToTime(value interface{}) (time.Time, error) {
    strValue, ok := value.(string)
    if !ok {
        return time.Time{}, fmt.Errorf("veri string formatinda değil")
    }

    unixTime, err := strconv.ParseInt(strValue, 10, 64)
    if err != nil {
        return time.Time{}, fmt.Errorf("veri int64 formatina çevrilemedi: %v", err)
    }

    return time.Unix(unixTime, 0), nil
}